// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"io/ioutil"
	"log"

	"github.com/edgelesssys/ego/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func main() {
	var t test.T
	defer t.Exit()

	assert := assert.New(&t)
	require := require.New(&t)

	log.Println("Welcome to the enclave.")
	testFileSystemMounts(assert, require)
}

func testFileSystemMounts(assert *assert.Assertions, require *require.Assertions) {
	log.Println("Testing hostfs mounts...")
	fileContent, err := ioutil.ReadFile("/data/test-file.txt")
	require.NoError(err)
	assert.Equal("It works!", string(fileContent))

	log.Println("Testing memfs mounts...")
	err = ioutil.WriteFile("/memfs/test-file.txt", fileContent, 0)
	require.NoError(err)
	newFileContent, err := ioutil.ReadFile("/memfs/test-file.txt")
	assert.Equal("It works!", string(newFileContent))
}
