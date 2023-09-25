// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"debug/elf"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"ego/config"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// create an unsigned EGo executable
var elfUnsigned = func() []byte {
	const outFile = "hello"
	const srcFile = outFile + ".go"

	goroot, err := filepath.Abs(filepath.Join("..", "..", "_ertgo"))
	if err != nil {
		panic(err)
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	// write minimal source file
	const src = `package main;import _"time";func main(){}`
	if err := ioutil.WriteFile(filepath.Join(dir, srcFile), []byte(src), 0o400); err != nil {
		panic(err)
	}

	// compile
	cmd := exec.Command(filepath.Join(goroot, "bin", "go"), "build", srcFile)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOROOT="+goroot)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	// read resulting executable
	data, err := ioutil.ReadFile(filepath.Join(dir, outFile))
	if err != nil {
		panic(err)
	}

	return data
}()

func TestEmbedConfigAsPayload(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	// Setup test environment
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	runner := signRunner{fs: fs}
	cli := NewCli(&runner, fs)

	// Copy from hostfs to memfs
	const exe = "helloworld"
	require.NoError(fs.WriteFile(exe, elfUnsigned, 0))

	// Check if no payload exists
	unsignedExeMemfs, err := fs.Open(exe)
	require.NoError(err)
	unsignedExeMemfsStat, err := unsignedExeMemfs.Stat()
	require.NoError(err)
	unsignedExeMemfsSize := unsignedExeMemfsStat.Size()
	payloadSize, payloadOffset, oeInfoOffset, err := getPayloadInformation(unsignedExeMemfs)
	assert.NoError(err)
	assert.Zero(payloadSize)
	assert.Zero(payloadOffset)
	assert.NotZero(oeInfoOffset)
	unsignedExeMemfs.Close()

	// Create a default config we want to check
	testConf := &config.Config{
		Exe:             exe,
		Key:             defaultPrivKeyFilename,
		Debug:           true,
		HeapSize:        512, //[MB]
		ProductID:       1,
		SecurityVersion: 1,
	}

	// Marshal config
	jsonData, err := json.Marshal(testConf)
	require.NoError(err)
	expectedLengthOfPayload := len(jsonData)

	// Embed json config to helloworld
	err = cli.embedConfigAsPayload(exe, jsonData)
	assert.NoError(err)

	// Check if new helloworld_signed now contains signed data
	signedExeMemfs, err := fs.Open(exe)
	require.NoError(err)
	defer signedExeMemfs.Close()

	payloadSize, payloadOffset, oeInfoOffset, err = getPayloadInformation(signedExeMemfs)
	assert.NoError(err)
	assert.EqualValues(expectedLengthOfPayload, payloadSize)
	assert.EqualValues(unsignedExeMemfsSize, payloadOffset)
	assert.NotZero(oeInfoOffset)

	// Reconstruct the JSON and check equality
	reconstructedJSON := make([]byte, payloadSize)
	n, err := signedExeMemfs.ReadAt(reconstructedJSON, payloadOffset)
	require.NoError(err)
	require.EqualValues(payloadSize, n)
	assert.EqualValues(jsonData, reconstructedJSON)

	// Now modify the config, redo everything and see if trucate worked fine and everything still lines up
	testConf.HeapSize = 5120
	jsonNewData, err := json.Marshal(testConf)
	require.NoError(err)
	expectedLengthOfNewPayload := len(jsonNewData)

	// Re-sign the already signed executable
	err = cli.embedConfigAsPayload(exe, jsonNewData)
	assert.NoError(err)
	payloadSize, payloadOffset, _, err = getPayloadInformation(signedExeMemfs)
	require.NoError(err)
	assert.EqualValues(expectedLengthOfNewPayload, payloadSize)
	assert.EqualValues(unsignedExeMemfsSize, payloadOffset)

	// Reconstruct the JSON and check if it not the old one, but the new one
	reconstructedJSON = make([]byte, payloadSize)
	n, err = signedExeMemfs.ReadAt(reconstructedJSON, payloadOffset)
	require.NoError(err)
	require.EqualValues(payloadSize, n)

	// Finally, check if we got the new JSON config and not the old one
	assert.NotEqualValues(jsonData, reconstructedJSON)
	assert.EqualValues(jsonNewData, reconstructedJSON)
}

func TestCheckHeapMode(t *testing.T) {
	testCases := map[string]struct {
		symbols  []elf.Symbol
		heapSize int
		want     error
	}{
		"default heap, small": {
			symbols:  []elf.Symbol{{Name: "runtime.arenaBaseOffset"}},
			heapSize: 511,
			want:     nil,
		},
		"default heap, lower bound": {
			symbols:  []elf.Symbol{{Name: "runtime.arenaBaseOffset"}},
			heapSize: 512,
			want:     nil,
		},
		"default heap, upper bound": {
			symbols:  []elf.Symbol{{Name: "runtime.arenaBaseOffset"}},
			heapSize: 16384,
			want:     nil,
		},
		"default heap, large": {
			symbols:  []elf.Symbol{{Name: "runtime.arenaBaseOffset"}},
			heapSize: 16385,
			want:     ErrNoLargeHeapWithLargeHeapSize,
		},
		"large heap, small": {
			heapSize: 511,
			want:     ErrLargeHeapWithSmallHeapSize,
		},
		"large heap, lower bound": {
			heapSize: 512,
			want:     nil,
		},
		"large heap, upper bound": {
			heapSize: 16384,
			want:     nil,
		},
		"large heap, large": {
			heapSize: 16385,
			want:     nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.want, checkHeapMode(tc.symbols, tc.heapSize))
		})
	}
}
