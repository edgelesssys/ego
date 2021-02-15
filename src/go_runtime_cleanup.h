// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

#pragma once

#include <openenclave/bits/defs.h>
#include <pthread.h>
#include <sys/mman.h>

OE_EXTERNC_BEGIN

/**
 * Creates a thread and adds it to the cleanup list.
 */
int go_rc_pthread_create(
    pthread_t* thread,
    const pthread_attr_t* attr,
    void* (*start_routine)(void*),
    void* arg);

/**
 * Maps a memory range and adds it to the cleanup list.
 */
void* go_rc_mmap(
    void* addr,
    size_t length,
    int prot,
    int flags,
    int fd,
    off_t offset);

/**
 * Unmaps a mapped memory range and removes it from the cleanup list.
 *
 * @param addr Start address of the memory mapping.
 * @param length Length of the memory mapping in bytes.
 */
int go_rc_munmap(void* addr, size_t length);

/**
 * Cancels all threads in the cleanup list.
 */
void go_rc_kill_threads();

/**
 * Unmaps all memory in the cleanup list.
 */
void go_rc_unmap_memory();

OE_EXTERNC_END
