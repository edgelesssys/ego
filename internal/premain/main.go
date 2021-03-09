// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import "C"

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/edgelesssys/ego/internal/premain/core"
)

var cargs []*C.char

// SyscallMounter uses an Open Enclave syscall to mount the filesystems in the Premain
type SyscallMounter struct{}

func main() {}

//export ert_ego_premain
func ert_ego_premain(argc *C.int, argv ***C.char, payload *C.char) **C.char {
	goNewEnviron, err := core.PreMain(C.GoString(payload), &SyscallMounter{})
	if err != nil {
		panic(err)
	}

	cargs = make([]*C.char, len(os.Args)+1)
	for i, a := range os.Args {
		cargs[i] = C.CString(a)
	}

	*argc = C.int(len(os.Args))
	*argv = &cargs[0]

	// Build and return new environ to use
	cNewEnviron := C.malloc(C.size_t(len(goNewEnviron)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	cNewEnvironIndexable := (*[1<<30 - 1]*C.char)(cNewEnviron)

	for index, value := range goNewEnviron {
		cNewEnvironIndexable[index] = C.CString(value)
	}

	return (**C.char)(cNewEnviron)
}

// Mount for SyscallMounter redirects to syscall.Mount
func (m *SyscallMounter) Mount(source string, target string, filesystem string, flags uintptr, data string) error {
	return syscall.Mount(source, target, filesystem, flags, data)
}
