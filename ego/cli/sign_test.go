// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"ego/config"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSign(t *testing.T) {
	const exefile = "exefile"

	testCases := map[string]struct {
		filename       string
		keyfilename    string
		existingFiles  map[string]string
		expectErr      bool
		expectedConfig string
		expectedKey    string
	}{
		"enclave.json does not exist": {
			expectErr: true,
		},
		"enclave.json is empty": {
			existingFiles: map[string]string{"enclave.json": ""},
			expectErr:     true,
		},
		"enclave.json is invalid": {
			existingFiles: map[string]string{"enclave.json": "foo"},
			expectErr:     true,
		},
		"empty json": {
			existingFiles: map[string]string{"enclave.json": "{}"},
			expectErr:     true,
		},
		"missing exe": {
			existingFiles: map[string]string{"enclave.json": `{"key":"keyfile", "heapSize":2}`},
			expectErr:     true,
		},
		"missing key": {
			existingFiles: map[string]string{"enclave.json": `{"exe":"exefile", "heapSize":2}`},
			expectErr:     true,
		},
		"missing heapSize": {
			existingFiles: map[string]string{"enclave.json": `{"exe":"exefile", "key":"keyfile"}`},
			expectErr:     true,
		},
		"key does not exist": {
			keyfilename:   "keyfile",
			existingFiles: map[string]string{"enclave.json": `{"exe":"exefile", "key":"keyfile", "heapSize":2}`},
			expectedConfig: `ProductID=0
SecurityVersion=0
Debug=0
NumHeapPages=512
NumStackPages=1024
NumTCS=32
`,
		},
		"key exists": {
			keyfilename: "keyfile",
			existingFiles: map[string]string{
				"keyfile":      "existingkey",
				"enclave.json": `{"exe":"exefile", "key":"keyfile", "heapSize":3, "debug":true, "productID":4, "securityVersion":5}`,
			},
			expectedConfig: `ProductID=4
SecurityVersion=5
Debug=1
NumHeapPages=768
NumStackPages=1024
NumTCS=32
`,
			expectedKey: "existingkey",
		},
		"key does not exist + custom config name": {
			filename:      "customConfigName.json",
			keyfilename:   "keyfile",
			existingFiles: map[string]string{"customConfigName.json": `{"exe":"exefile", "key":"keyfile", "heapSize":2}`},
			expectedConfig: `ProductID=0
SecurityVersion=0
Debug=0
NumHeapPages=512
NumStackPages=1024
NumTCS=32
`,
		},
		"files not in working dir": {
			filename:      "foo/enclave.json",
			keyfilename:   "bar/keyfile",
			existingFiles: map[string]string{"foo/enclave.json": `{"exe":"../exefile", "key":"../bar/keyfile", "heapSize":2}`},
			expectedConfig: `ProductID=0
SecurityVersion=0
Debug=0
NumHeapPages=512
NumStackPages=1024
NumTCS=32
`,
		},
		"sign executable": {
			filename:    exefile,
			keyfilename: "private.pem",
			expectedConfig: `ProductID=1
SecurityVersion=1
Debug=1
NumHeapPages=131072
NumStackPages=1024
NumTCS=32
`,
		},
		"exe in enclave.json does not match provided exefile": {
			filename:      "notExefile",
			existingFiles: map[string]string{"enclave.json": `{"exe":"exefile", "key":"keyfile", "heapSize":2}`},
			expectErr:     true,
		},
		"exe in enclave.json matches provided exefile": {
			filename:      exefile,
			keyfilename:   "keyfile",
			existingFiles: map[string]string{"enclave.json": `{"exe":"exefile", "key":"keyfile", "heapSize":2}`},
			expectedConfig: `ProductID=0
SecurityVersion=0
Debug=0
NumHeapPages=512
NumStackPages=1024
NumTCS=32
`,
		},
		"executable heap": {
			keyfilename:   "keyfile",
			existingFiles: map[string]string{"enclave.json": `{"exe":"exefile", "key":"keyfile", "heapSize":2, "executableHeap":true}`},
			expectedConfig: `ProductID=0
SecurityVersion=0
Debug=0
NumHeapPages=512
NumStackPages=1024
NumTCS=32
ExecutableHeap=1
`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			runner := signRunner{fs: fs, expectedConfig: tc.expectedConfig}
			cli := NewCli(&runner, fs)

			// Create files
			require.NoError(fs.WriteFile(exefile, elfUnsigned, 0))
			for name, data := range tc.existingFiles {
				require.NoError(fs.WriteFile(name, []byte(data), 0))
			}

			// Perform the signing
			err := cli.Sign(tc.filename)
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			// Check private and public key
			key, err := fs.ReadFile(tc.keyfilename)
			require.NoError(err)
			exists, err := fs.Exists(filepath.Join(filepath.Dir(tc.keyfilename), "public.pem"))
			if tc.expectedKey == "" {
				assert.EqualValues("newkey", key)
				require.NoError(err)
				assert.True(exists)
			} else {
				assert.EqualValues(tc.expectedKey, key)
				require.NoError(err)
				assert.False(exists)
			}
		})
	}
}

func TestEmbedFile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	runner := signRunner{fs: fs}
	cli := NewCli(&runner, fs)

	// Create executable
	const exe = "exefile"
	require.NoError(fs.WriteFile(exe, elfUnsigned, 0))

	// Create file to embed
	const sourcePath = "/src"
	const targetPath = "/path/to/file"
	content := []byte{2, 0, 3}
	require.NoError(fs.WriteFile(sourcePath, content, 0))

	// Sign executable, which should embed the file
	const configJSON = `
{
	"exe": "` + exe + `",
	"key": "private.pem",
	"heapSize": 1,
	"files": [
		{
			"source": "` + sourcePath + `",
			"target": "` + targetPath + `"
		}
	]
}
`
	runner.expectedConfig = `ProductID=0
SecurityVersion=0
Debug=0
NumHeapPages=256
NumStackPages=1024
NumTCS=32
`
	require.NoError(fs.WriteFile("enclave.json", []byte(configJSON), 0))
	require.NoError(cli.Sign(""))

	// Get payload
	file, err := fs.Open(exe)
	require.NoError(err)
	defer file.Close()
	payloadSize, payloadOffset, _, err := getPayloadInformation(file)
	require.NoError(err)
	payload := make([]byte, payloadSize)
	_, err = file.ReadAt(payload, payloadOffset)
	require.NoError(err)

	// Verify that config includes embedded file
	var conf config.Config
	require.NoError(json.Unmarshal(payload, &conf))
	require.Len(conf.Files, 1)
	assert.Equal(sourcePath, conf.Files[0].Source)
	assert.Equal(targetPath, conf.Files[0].Target)
	actualContent, err := conf.Files[0].GetContent()
	require.NoError(err)
	require.Equal(content, actualContent)
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
		cmp.Equal(cmd.Args[4:6], []string{"-pubout", "-out"}) {
		exists, err := s.fs.Exists(cmd.Args[3])
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("openssl rsa: " + cmd.Args[3] + " does not exist")
		}
		return s.fs.WriteFile(cmd.Args[6], nil, 0)
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
