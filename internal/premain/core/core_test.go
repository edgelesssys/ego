// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall"
	"testing"

	"github.com/edgelesssys/ego/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type assertionMounter struct {
	assert      *assert.Assertions
	config      *config.Config
	usedTargets map[string]bool
}

func TestPremain(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	//sane default values
	conf := &config.Config{
		Exe:             "helloworld",
		Key:             "privatekey",
		Debug:           true,
		HeapSize:        512, //[MB]
		ProductID:       1,
		SecurityVersion: 1,
		Mounts:          []config.FileSystemMount{{Source: "/", Target: "/memfs", Type: "memfs", ReadOnly: false}, {Source: "/home/benjaminfranklin", Target: "/data", Type: "hostfs", ReadOnly: true}},
	}

	// Supply valid payload, no Marble
	mounter := assertionMounter{assert: assert, config: conf, usedTargets: make(map[string]bool)}
	_, err := PreMain("", &mounter)
	assert.NoError(err)
	// Supply valid payload, no Marble
	payload, err := json.Marshal(conf)
	require.NoError(err)
	mounter = assertionMounter{assert: assert, config: conf, usedTargets: make(map[string]bool)}
	_, err = PreMain(string(payload), &mounter)
	assert.NoError(err)
	// Supply invalid payload, should fail
	payload = []byte("blablarubbish")
	_, err = PreMain(string(payload), &mounter)
	assert.Error(err)
}

func TestPerformMounts(t *testing.T) {
	assert := assert.New(t)

	//sane default values
	conf := &config.Config{
		Exe:             "helloworld",
		Key:             "privatekey",
		Debug:           true,
		HeapSize:        512, //[MB]
		ProductID:       1,
		SecurityVersion: 1,
		Mounts:          []config.FileSystemMount{{Source: "/", Target: "/memfs", Type: "memfs", ReadOnly: false}, {Source: "/home/benjaminfranklin", Target: "/data", Type: "hostfs", ReadOnly: true}},
	}

	mounter := assertionMounter{assert: assert, config: conf, usedTargets: make(map[string]bool)}
	assert.NoError(performMounts(*conf, &mounter))

	conf.Mounts = []config.FileSystemMount{{Source: "/home/benjaminfranklin", Target: "/data", Type: "rubbishfs", ReadOnly: true}}
	assert.Error(performMounts(*conf, &mounter))
}

func (a *assertionMounter) Mount(source string, target string, filesystem string, flags uintptr, data string) error {
	// Find corresponding mount point in config by searching for the target
	var currentMountPoint config.FileSystemMount

	// Find target from config to check against
	// Additionally, check if it's already been mounted
	for _, mountPoint := range a.config.Mounts {
		if target == mountPoint.Target {
			if _, ok := a.usedTargets[mountPoint.Target]; ok {
				return errors.New("target already exists")
			}

			currentMountPoint = mountPoint
			break
		}
	}

	// We should not end up here, use this when we did not find any entry in the config
	if currentMountPoint.Type == "" {
		return errors.New("could not find equal mount declaration in supplied config")
	}

	if filesystem == "oe_host_file_system" {
		a.assert.EqualValues(currentMountPoint.Source, source)
		a.assert.EqualValues(currentMountPoint.Target, target)
		a.assert.EqualValues("hostfs", currentMountPoint.Type)
	} else if filesystem == "edg_memfs" {
		a.assert.EqualValues("/", source)
		a.assert.EqualValues(currentMountPoint.Target, target)
		a.assert.EqualValues("memfs", currentMountPoint.Type)
	} else {
		return errors.New("encountered a call to an unknown filesystem type")
	}

	if flags == syscall.MS_RDONLY {
		a.assert.True(currentMountPoint.ReadOnly)
	} else if flags == 0 {
		a.assert.False(currentMountPoint.ReadOnly)
	} else {
		return fmt.Errorf("unexpected flag supplied to mount: %d", flags)
	}

	a.assert.Empty(data)

	// Add to usedTargets list for duplication check
	a.usedTargets[currentMountPoint.Target] = true

	return nil
}
