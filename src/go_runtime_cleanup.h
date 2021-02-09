// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

#pragma once

#include <openenclave/bits/defs.h>
#include <pthread.h>
#include <stddef.h>
#include <stdint.h>

OE_EXTERNC_BEGIN

/**
 * Adds a thread from the cleanup list.
 *
 * @param thread pthread.
 */
void go_rc_add_thread(pthread_t thread);

/**
 * Adds a mapped memory range to the cleanup list.
 *
 * @param addr Start address of the memory mapping.
 * @param length Length of the memory mapping in bytes.
 */
void go_rc_add_memory(void* addr, uintptr_t length);

/**
 * Removes a mapped memory range from the cleanup list.
 *
 * @param addr Start address of the memory mapping.
 * @param length Length of the memory mapping in bytes.
 */
void go_rc_remove_memory(void* addr, uintptr_t length);

/**
 * Cancels all threads in the cleanup list.
 */
void go_rc_kill_threads();

/**
 * Unmaps all memory in the cleanup list.
 */
void go_rc_unmap_memory();

OE_EXTERNC_END
