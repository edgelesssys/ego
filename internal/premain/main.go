// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import "C"

import (
	"os"

	"github.com/edgelesssys/ego/internal/premain/core"
)

var cargs []*C.char

func main() {}

//export ert_ego_premain
func ert_ego_premain(argc *C.int, argv ***C.char, payload *C.char) {
	if err := core.PreMain(C.GoString(payload)); err != nil {
		panic(err)
	}

	cargs = make([]*C.char, len(os.Args)+1)
	for i, a := range os.Args {
		cargs[i] = C.CString(a)
	}

	*argc = C.int(len(os.Args))
	*argv = &cargs[0]
}
