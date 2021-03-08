// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	if err := ioutil.WriteFile(filepath.Join(dir, srcFile), []byte(src), 0400); err != nil {
		panic(err)
	}

	// compile
	cmd := exec.Command(filepath.Join(goroot, "bin", "go"), "build", srcFile)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOROOT="+goroot)
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

func TestSignEmptyFilename(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	runner := signRunner{fs: fs}
	cli := NewCli(&runner, fs)

	// Create executable
	require.NoError(fs.WriteFile("exefile", elfUnsigned, 0))

	// enclave.json does not exist
	require.Error(cli.Sign(""))

	// enclave.json is empty
	require.NoError(fs.WriteFile("enclave.json", nil, 0))
	require.Error(cli.Sign(""))

	// enclave.json is invalid
	require.NoError(fs.WriteFile("enclave.json", []byte("foo"), 0))
	require.Error(cli.Sign(""))

	// empty json
	require.NoError(fs.WriteFile("enclave.json", []byte("{}"), 0))
	require.Error(cli.Sign(""))

	// missing exe
	require.NoError(fs.WriteFile("enclave.json", []byte(`{"key":"keyfile", "heapSize":2}`), 0))
	require.Error(cli.Sign(""))

	// missing key
	require.NoError(fs.WriteFile("enclave.json", []byte(`{"exe":"exefile", "heapSize":2}`), 0))
	require.Error(cli.Sign(""))

	// missing heapSize
	require.NoError(fs.WriteFile("enclave.json", []byte(`{"exe":"exefile", "key":"keyfile"}`), 0))
	require.Error(cli.Sign(""))

	// key does not exist
	runner.expectedConfig = `ProductID=0
SecurityVersion=0
Debug=0
NumHeapPages=512
NumStackPages=1024
NumTCS=32
`
	require.NoError(fs.WriteFile("enclave.json", []byte(`{"exe":"exefile", "key":"keyfile", "heapSize":2}`), 0))
	require.NoError(cli.Sign(""))
	key, err := fs.ReadFile("keyfile")
	require.NoError(err)
	assert.EqualValues("newkey", key)
	exists, err := fs.Exists("public.pem")
	require.NoError(err)
	assert.True(exists)

	// key exists
	runner.expectedConfig = `ProductID=4
SecurityVersion=5
Debug=1
NumHeapPages=768
NumStackPages=1024
NumTCS=32
`
	require.NoError(fs.WriteFile("keyfile", []byte("existingkey"), 0))
	require.NoError(fs.WriteFile("enclave.json", []byte(`{"exe":"exefile", "key":"keyfile", "heapSize":3, "debug":true, "productID":4, "securityVersion":5}`), 0))
	require.NoError(cli.Sign(""))
	key, err = fs.ReadFile("keyfile")
	require.NoError(err)
	assert.EqualValues("existingkey", key)
}

func TestSignConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	runner := signRunner{fs: fs}
	cli := NewCli(&runner, fs)

	// Create executable
	require.NoError(fs.WriteFile("exefile", elfUnsigned, 0))

	// key does not exist + custom config name
	runner.expectedConfig = `ProductID=0
SecurityVersion=0
Debug=0
NumHeapPages=512
NumStackPages=1024
NumTCS=32
`
	require.NoError(fs.WriteFile("customConfigName.json", []byte(`{"exe":"exefile", "key":"keyfile", "heapSize":2, "debug":false, "productID":0, "securityVersion":0}`), 0))
	require.NoError(cli.Sign("customConfigName.json"))
	key, err := fs.ReadFile("keyfile")
	require.NoError(err)
	assert.EqualValues("newkey", key)
	exists, err := fs.Exists("public.pem")
	require.NoError(err)
	assert.True(exists)

}

func TestSignExecutable(t *testing.T) {
	require := require.New(t)

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	runner := signRunner{fs: fs}
	cli := NewCli(&runner, fs)

	// Create executable
	require.NoError(fs.WriteFile("exefile", elfUnsigned, 0))

	// enclave.json does not exist
	runner.expectedConfig = `ProductID=1
SecurityVersion=1
Debug=1
NumHeapPages=131072
NumStackPages=1024
NumTCS=32
`
	require.NoError(cli.Sign("exefile"))

	//exe in enclave.json does not match provided exefile
	require.NoError(fs.WriteFile("enclave.json", []byte(`{"exe":"exefile", "key":"keyfile", "heapSize":3, "debug":true, "productID":4, "securityVersion":5}`), 0))
	require.Error(cli.Sign("notExefile"))

	//exe in enclave.json matches provided exefile
	runner.expectedConfig = `ProductID=4
SecurityVersion=5
Debug=1
NumHeapPages=768
NumStackPages=1024
NumTCS=32
`
	require.NoError(fs.WriteFile("enclave.json", []byte(`{"exe":"exefile", "key":"keyfile", "heapSize":3, "debug":true, "productID":4, "securityVersion":5}`), 0))
	require.NoError(cli.Sign("exefile"))
}

func TestSignJSONExecutablePayload(t *testing.T) {
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
	testConf := &config{
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
	payloadSize, payloadOffset, oeInfoOffset, err = getPayloadInformation(signedExeMemfs)
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

type signRunner struct {
	fs             afero.Afero
	expectedConfig string
}

func (s signRunner) Run(cmd *exec.Cmd) error {
	if cmp.Equal(cmd.Args[:3], []string{"openssl", "genrsa", "-out"}) &&
		cmp.Equal(cmd.Args[4:], []string{"-3", "3072"}) {
		return s.fs.WriteFile(cmd.Args[3], []byte("newkey"), 0)
	}
	if cmp.Equal(cmd.Args[:3], []string{"openssl", "rsa", "-in"}) &&
		cmp.Equal(cmd.Args[4:], []string{"-pubout", "-out", "public.pem"}) {
		exists, err := s.fs.Exists(cmd.Args[3])
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("openssl rsa: " + cmd.Args[3] + " does not exist")
		}
		return s.fs.WriteFile("public.pem", nil, 0)
	}
	return errors.New("unexpected cmd: " + cmd.Path)
}

func (signRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	panic(cmd.Path)
}

func (s signRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	if !(filepath.Base(cmd.Path) == "ego-oesign" &&
		cmp.Equal(cmd.Args[1:3], []string{"sign", "-e"}) &&
		cmp.Equal(cmd.Args[6], "-k") &&
		cmp.Equal(cmd.Args[8:], []string{"--payload", "exefile"})) {
		return nil, errors.New("unexpected cmd: " + cmd.Path + strings.Join(cmd.Args, " "))
	}
	data, err := s.fs.ReadFile(cmd.Args[5])
	if err != nil {
		return nil, err
	}
	config := string(data)
	if config != s.expectedConfig {
		return nil, errors.New("unexpected config: " + config)
	}
	return nil, nil
}

func (signRunner) ExitCode(cmd *exec.Cmd) int {
	panic(cmd.Path)
}
