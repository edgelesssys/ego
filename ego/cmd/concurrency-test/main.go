// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"runtime"
	"syscall"
)

// This tests the EGOMAXTHREADS feature for scheduling issues.

func main() {
	const (
		numTCS     = 9            // sync with enclave.conf
		maxProcs   = numTCS + 2   // greater than TCS so that it would cause OE_OUT_OF_THREADS with the unmodified Go scheduler
		numWorkers = 3 * maxProcs // greater than maxProcs so that workers must be preempted
	)

	oldMaxProcs := runtime.GOMAXPROCS(maxProcs)
	if !(2 <= oldMaxProcs && oldMaxProcs <= numTCS-4) { // EGo enforces default GOMAXPROCS to be in this range
		panic(fmt.Sprintf("oldMaxProcs is %v", oldMaxProcs))
	}

	test(numWorkers)
}

func test(numWorkers int) {
	// Create some Go routines that are connected by channels and pass values from the first to the last.
	// If any of the Go routines wouldn't be scheduled anymore, the others couldn't make progress either.
	c0 := make(chan int)
	c1 := c0
	for i := 0; i < numWorkers; i++ {
		c2 := make(chan int)
		go work(c1, c2)
		c1 = c2
	}

	go produce(c0)
	sum := 0
	for x := range c1 {
		sum += x
	}
	fmt.Println(sum)
}

func produce(ch chan int) {
	const count = 10_000
	sum := 0
	for i := 0; i < count; i++ {
		if i%(count/10) == 0 {
			fmt.Printf("%v / %v\n", i, count)
		}
		sum += i
		ch <- i
	}
	fmt.Println(sum)
	close(ch)
}

func work(in, out chan int) {
	var garbage []byte
	for x := range in {
		out <- x

		// cgo call triggers specific scheduling logic
		_ = syscall.Nanosleep(&syscall.Timespec{Sec: 0, Nsec: 1_000}, nil)

		// triggers GC from time to time
		garbage = make([]byte, 1024)
	}
	_ = garbage
	close(out)
}
