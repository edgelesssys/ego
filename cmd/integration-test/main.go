// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"

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
	testEnvVars(assert, require)
}

func testFileSystemMounts(assert *assert.Assertions, require *require.Assertions) {
	// Check if root directory is empty
	log.Println("Testing default memfs mount...")
	log.Println("Check if exposed filesystem only contains 'edg' for memfs mounts...")
	dirContent, err := os.Open("/")
	require.NoError(err)
	content, err := dirContent.Readdirnames(2)
	assert.NoError(err)
	assert.Len(content, 1)
	assert.Equal("edg", content[0])

	// Check if we can write and read to the root memfs
	log.Println("Checking I/O of memfs...")
	const localTest = "This is a test!"
	require.NoError(ioutil.WriteFile("test-root.txt", []byte(localTest), 0755))
	fileContent, err := ioutil.ReadFile("test-root.txt")
	require.NoError(err)
	assert.Equal(localTest, string(fileContent))

	// Check hostfs mounts specified in manifest
	log.Println("Testing hostfs mounts...")
	fileContent, err = ioutil.ReadFile("/data/test-file.txt")
	require.NoError(err)
	assert.Equal("It works!", string(fileContent))

	// Check memfs mounts specified in manifest
	log.Println("Testing memfs mounts...")

	// Check if new memfs mount does not contain any files from the root filesystem
	dirContent, err = os.Open("/memfs")
	require.NoError(err)
	_, err = dirContent.Readdirnames(1)
	assert.ErrorIs(io.EOF, err)

	// Check if we can write and read to the mounted memfs
	err = ioutil.WriteFile("/memfs/test-file.txt", fileContent, 0)
	require.NoError(err)
	newFileContent, err := ioutil.ReadFile("/memfs/test-file.txt")
	require.NoError(err)
	assert.Equal("It works!", string(newFileContent))

	// Check if we can read to the mounted memfs from root explicitly
	newFileContent, err = ioutil.ReadFile("/edg/mnt/memfs/test-file.txt")
	require.NoError(err)
	assert.Equal("It works!", string(newFileContent))
}

func testEnvVars(assert *assert.Assertions, require *require.Assertions) {
	// Test if new env vars were set
	log.Println("Testing env vars...")
	assert.Equal("Let's hope this passes the test :)", os.Getenv("HELLO_WORLD"))
	currentPwd, err := os.Getwd()
	require.NoError(err)
	assert.Equal(currentPwd, "/data")

	// Test if OE_IS_ENCLAVE is set
	assert.Equal("1", os.Getenv("OE_IS_ENCLAVE"))

	// Check if other env vars were not taken over by using a common one to check against (here: LANG)
	assert.Empty(os.Getenv("LANG"))
}
