// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
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

	// Specify source path for memfs & set it as read only. Really makes no sense and throws warnings, but should pass.
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
