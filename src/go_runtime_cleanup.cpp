// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

#include "go_runtime_cleanup.h"
#include <openenclave/advanced/allocator.h>
#include <openenclave/ert.h>
#include <sys/mman.h>
#include <algorithm>
#include <cassert>
#include <climits>
#include <cstdlib>
#include <cstring>
#include <iostream>
#include <mutex>
#include <stdexcept>
#include <vector>
#include "bitset.h"

using namespace std;

static vector<pthread_t> threads;
static mutex _mux;
static void* _bitset;
static size_t _bitmap_size;
static const void* _base;

extern "C" void* __oe_get_heap_base();
extern "C" size_t __oe_get_heap_size();
extern "C" int oe_epoll_wake();

uint64_t _oe_round_up_to_page_size(uint64_t x)
{
    uint64_t n = OE_PAGE_SIZE;
    return (x + n - 1) / n * n;
}

static size_t _to_pos(const void* addr)
{
    return (size_t)((uint8_t*)addr - (uint8_t*)_base) / OE_PAGE_SIZE;
}

static const void* _to_addr(size_t pos)
{
    return (const void*)((uint8_t*)_base + (pos * OE_PAGE_SIZE));
}

static const int _go_rc_init = [] {
    _base = __oe_get_heap_base();
    size_t heap_size = __oe_get_heap_size();
    _bitmap_size = heap_size / OE_PAGE_SIZE;
    size_t alloc_size = _oe_round_up_to_page_size(_bitmap_size / CHAR_BIT);
    _bitset = oe_allocator_calloc(1, alloc_size);
    return 1;
}();

int go_rc_pthread_create(
    pthread_t* thread,
    const pthread_attr_t* attr,
    void* (*start_routine)(void*),
    void* arg)
{
    const lock_guard<mutex> lock(_mux);
    const int res = pthread_create(thread, attr, start_routine, arg);
    if (res == 0)
        threads.push_back(*thread);
    return res;
}

void* go_rc_mmap(
    void* addr,
    size_t length,
    int prot,
    int flags,
    int fd,
    off_t offset)
{
    const lock_guard<mutex> lock(_mux);
    const auto res = mmap(addr, length, prot, flags, fd, offset);
    if (res != MAP_FAILED)
        ert_bitset_set_range(_bitset, _to_pos(res), length / OE_PAGE_SIZE);
    return res;
}

int go_rc_munmap(void* addr, size_t length)
{
    const lock_guard<mutex> lock(_mux);
    const int res = munmap(addr, length);
    if (res == 0)
        ert_bitset_reset_range(_bitset, _to_pos(addr), length / OE_PAGE_SIZE);
    return res;
}

void go_rc_kill_threads()
{
    const lock_guard<mutex> lock(_mux);
    for (const auto thread : threads)
    {
        const int ret = pthread_cancel(thread);
        if (ret != 0)
        {
            errno = ret;
            perror("pthread_cancel");
            return;
        }
    }
    const int ret = oe_epoll_wake();
    if (ret != 0)
    {
        errno = ret;
        perror("oe_epoll_wake");
        return;
    }
    for (const auto thread : threads)
    {
        pthread_join(thread, nullptr);
    }

    threads.clear();
}

void go_rc_unmap_memory()
{
    for (size_t pos = 0;;)
    {
        size_t pages = 0;
        pos = ert_bitset_find_set_range(_bitset, _bitmap_size, pos, &pages);
        if (pos == SIZE_MAX)
        {
            return;
        }
        const void* addr = (void*)_to_addr(pos);
        size_t length = pages * OE_PAGE_SIZE;
        munmap((void*)addr, length);
        ert_bitset_reset_range(_bitset, pos, pages);
    }
}
