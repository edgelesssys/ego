// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import "C"

import (
	"ego/premain/core"
	"os"
	"syscall"
	"unsafe"

	"github.com/spf13/afero"
)

var cargs []*C.char

// SyscallMounter uses an Open Enclave syscall to mount the filesystems in the Premain
type SyscallMounter struct{}

func main() {}

//export ert_ego_premain
func ert_ego_premain(argc *C.int, argv ***C.char, envc C.int, envp **C.char, payload *C.char) {
	originalEnviron := convertEnvironmentToGoStringArray(envc, envp)
	if err := core.PreMain(C.GoString(payload), &SyscallMounter{}, afero.NewOsFs(), originalEnviron); err != nil {
		panic(err)
	}

	cargs = make([]*C.char, len(os.Args)+1)
	for i, a := range os.Args {
		cargs[i] = C.CString(a)
	}

	*argc = C.int(len(os.Args))
	*argv = &cargs[0]
}

// Mount for SyscallMounter redirects to syscall.Mount
func (m *SyscallMounter) Mount(source string, target string, filesystem string, flags uintptr, data string) error {
	return syscall.Mount(source, target, filesystem, flags, data)
}

// Unmount for SyscallMounter redirects to syscall.Unmount
func (m *SyscallMounter) Unmount(target string, flags int) error {
	return syscall.Unmount(target, flags)
}

// Adapted from: https://stackoverflow.com/a/36189294
func convertEnvironmentToGoStringArray(envc C.int, envp **C.char) []string {
	length := int(envc)
	tmpSlice := (*[1 << 30]*C.char)(unsafe.Pointer(envp))[:length:length]
	goStrings := make([]string, length)
	for i, s := range tmpSlice {
		goStrings[i] = C.GoString(s)
	}
	return goStrings
}
