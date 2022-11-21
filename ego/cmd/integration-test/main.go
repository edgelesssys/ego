// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"ego/test"
	"io"
	"log"
	"os"

	"github.com/klauspost/cpuid/v2"
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
	testCpuid(assert, require)
}

func testFileSystemMounts(assert *assert.Assertions, require *require.Assertions) {
	// Check if root directory is empty
	log.Println("Testing default memfs mount...")
	dirContent, err := os.Open("/")
	require.NoError(err)
	content, err := dirContent.Readdirnames(0)
	assert.NoError(err)
	assert.Len(content, 2)
	assert.Contains(content, "edg")
	assert.Contains(content, "path")

	// Check if we can write and read to the root memfs
	log.Println("Checking I/O of memfs...")
	const localTest = "This is a test!"
	require.NoError(os.WriteFile("test-root.txt", []byte(localTest), 0o755))
	fileContent, err := os.ReadFile("test-root.txt")
	require.NoError(err)
	assert.Equal(localTest, string(fileContent))

	// Check hostfs mounts specified in manifest
	log.Println("Testing hostfs mounts...")
	fileContent, err = os.ReadFile("/reldata/test-file.txt")
	require.NoError(err)
	assert.Equal("It relatively works!", string(fileContent))
	fileContent, err = os.ReadFile("/data/test-file.txt")
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
	err = os.WriteFile("/memfs/test-file.txt", fileContent, 0o400)
	require.NoError(err)
	newFileContent, err := os.ReadFile("/memfs/test-file.txt")
	require.NoError(err)
	assert.Equal("It works!", string(newFileContent))

	// Check if we can read to the mounted memfs from root explicitly
	newFileContent, err = os.ReadFile("/edg/mnt/memfs/test-file.txt")
	require.NoError(err)
	assert.Equal("It works!", string(newFileContent))

	// Check embedded file
	buff, err := os.ReadFile("/path/to/file_enclave.txt")
	require.NoError(err)
	assert.Equal("i should be in memfs", string(buff))
	require.NoError(os.WriteFile("/path/to/file_enclave.txt", []byte{2}, 0))
}

func testEnvVars(assert *assert.Assertions, require *require.Assertions) {
	// Test if new env vars were set
	log.Println("Testing env vars...")
	assert.Equal("Let's hope this passes the test :)", os.Getenv("HELLO_WORLD"))
	currentPwd, err := os.Getwd()
	require.NoError(err)
	assert.Equal("/data", currentPwd)

	// Test if OE_IS_ENCLAVE is set
	assert.Equal("1", os.Getenv("OE_IS_ENCLAVE"))

	// Check if other env vars were not taken over by using a common one to check against (here: LANG)
	assert.Empty(os.Getenv("LANG"))
}

func testCpuid(assert *assert.Assertions, require *require.Assertions) {
	assert.True(cpuid.CPU.Has(cpuid.CMOV))
}
