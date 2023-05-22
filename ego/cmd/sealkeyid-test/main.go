// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"

	"github.com/edgelesssys/ego/enclave"
)

func main() {
	id, err := enclave.GetSealKeyID()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%x\n", id)
}
