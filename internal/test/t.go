// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package test

import (
	"fmt"
	"os"
	"runtime"
)

// T implements the interface required by testify.
type T struct {
	exitCode int
}

// Errorf prints an error message and marks the test as having failed.
func (t *T) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	t.exitCode = 1
}

// FailNow exits the program with an error. You must call `defer t.Exit()` first.
func (t *T) FailNow() {
	t.exitCode = 1
	runtime.Goexit()
}

// Exit exits the program with an appropriate exit code.
func (t *T) Exit() {
	var msg string
	if t.exitCode == 0 {
		msg = "passed"
	} else {
		msg = "failed"
	}
	fmt.Println(msg)
	os.Exit(t.exitCode)
}
