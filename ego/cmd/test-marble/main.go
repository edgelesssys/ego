// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"os"

	"github.com/edgelesssys/ego/ego/test"
	"github.com/stretchr/testify/assert"
)

func main() {
	var t test.T
	defer t.Exit()
	assert := assert.New(&t)

	assert.Equal([]string{"arg0", "arg1", "arg2"}, os.Args)
	assert.Equal("val1", os.Getenv("key1"))
	assert.Equal("val2", os.Getenv("key2"))
}
