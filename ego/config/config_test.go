// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMinimalConfig(t *testing.T) {
	assert := assert.New(t)

	config := Config{}

	// Empty config, should fail
	assert.Error(config.Validate())

	// Set heapSize, should still fail
	config.HeapSize = 512
	assert.Error(config.Validate())

	// Set exe, should still fail
	config.Exe = "text_exe"
	assert.Error(config.Validate())

	// Set key, should pass now
	config.Key = "somekey.key"
	assert.NoError(config.Validate())
}

func TestValidateFileSystemMounts(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	config := Config{}

	// Set minimal parameters for config
	config.HeapSize = 512
	config.Exe = "text_exe"
	config.Key = "somekey.key"
	require.NoError(config.Validate())

	// Set two valid mount options, should pass
	config.Mounts = make([]FileSystemMount, 2)

	// Run with empty mounts, should fail
	assert.Error(config.Validate())

	// Run with empty type, should fail
	config.Mounts[0] = FileSystemMount{Source: "/", Target: "/data_memfs", ReadOnly: false}
	assert.Error(config.Validate())

	// Run with proper definitions for mounts, should pass
	config.Mounts[0] = FileSystemMount{Source: "/", Target: "/data_memfs", Type: "memfs", ReadOnly: false}
	config.Mounts[1] = FileSystemMount{Source: "/home/benjaminfranklin", Target: "/data", Type: "hostfs", ReadOnly: false}
	assert.NoError(config.Validate())

	// Specify source path for memfs, should pass.
	config.Mounts[0] = FileSystemMount{Source: "/blabla", Target: "/data_memfs", Type: "memfs", ReadOnly: true}
	assert.NoError(config.Validate())

	// Specify no source hostfs, should fail
	config.Mounts[0] = FileSystemMount{Source: "", Target: "/sometarget", Type: "hostfs", ReadOnly: true}
	assert.Error(config.Validate())

	// Specify two mounts with /data_memfs as target, should fail
	config.Mounts[0] = FileSystemMount{Source: "/blabla", Target: "/data_memfs", Type: "memfs", ReadOnly: true}
	config.Mounts[1] = FileSystemMount{Source: "/home/benjaminfranklin", Target: "/data_memfs", Type: "hostfs", ReadOnly: false}
	assert.Error(config.Validate())

	// Specify garbage fs, should fail
	config.Mounts[0] = FileSystemMount{Source: "/makesNoSense", Target: "/bin", Type: "rubbishfs", ReadOnly: true}
	assert.Error(config.Validate())
}

func TestValidateEnvVars(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	config := Config{}

	// Set minimal parameters for config
	config.HeapSize = 512
	config.Exe = "text_exe"
	config.Key = "somekey.key"
	require.NoError(config.Validate())

	// Set an empty env var definition, should fail
	config.Env = make([]EnvVar, 1)
	assert.Error(config.Validate())

	// Set an valid env var definition, should pass
	config.Env[0] = EnvVar{Name: "HELLO_WORLD", Value: "1"}
	assert.NoError(config.Validate())

	// Set an valid env var definition with copy from host when existing, should pass
	config.Env[0] = EnvVar{Name: "HELLO_WORLD", Value: "1", FromHost: true}
	assert.NoError(config.Validate())

	// Add disallowed char to EnvVarName, should fail
	config.Env[0] = EnvVar{Name: "=HELLO_WORLD", Value: "1", FromHost: true}
	assert.Error(config.Validate())

	// Add env var with no value specified nor copying it from host, should trigger a warning and will not be added, but should pass
	config.Env[0] = EnvVar{Name: "HELLO_WORLD"}
	assert.NoError(config.Validate())

	// Add multiple entries with the same name, should fail
	config.Env = make([]EnvVar, 2)
	config.Env[0] = EnvVar{Name: "HELLO_WORLD", Value: "1", FromHost: true}
	config.Env[1] = EnvVar{Name: "HELLO_WORLD", Value: "1", FromHost: true}
	assert.Error(config.Validate())
}
func TestCustomFile(t *testing.T) {
	// valid base64
	// input == output
	assert := assert.New(t)
	assert.Nil(nil)
	require := require.New(t)
	require.Nil(nil)

	config := Config{Files: []File{{Source: "testfile.txt", Target: "/dir/testfile_target.txt"}}}
	err := config.PopulateContent()
	require.NoError(err)
	buf, err := config.Files[0].GetContent()
	require.NoError(err)
	assert.Equal("just some string", string(buf))

}
