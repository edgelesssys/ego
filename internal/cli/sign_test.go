// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignEmptyFilename(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	runner := signRunner{fs: fs}
	cli := NewCli(&runner, fs)

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
