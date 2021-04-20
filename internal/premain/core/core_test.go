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
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/edgelesssys/ego/internal/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type assertionMounter struct {
	assert          *assert.Assertions
	config          *config.Config
	usedTargets     map[string]bool
	remountAsHostFS bool
}

func TestPremain(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	fs := afero.NewMemMapFs()

	//sane default values
	conf := &config.Config{
		Exe:             "helloworld",
		Key:             "privatekey",
		Debug:           true,
		HeapSize:        512, //[MB]
		ProductID:       1,
		SecurityVersion: 1,
		Mounts:          []config.FileSystemMount{{Source: "/", Target: "/memfs", Type: "memfs", ReadOnly: false}, {Source: "/home/benjaminfranklin", Target: "/data", Type: "hostfs", ReadOnly: true}},
		Env:             []config.EnvVar{{Name: "HELLO_WORLD", Value: "2"}, {Name: "PWD", Value: "/tmp/somedir", FromHost: true}},
	}

	// Supply valid payload, no Marble
	mounter := assertionMounter{assert: assert, config: conf, usedTargets: make(map[string]bool), remountAsHostFS: false}
	assert.NoError(PreMain("", &mounter, fs))

	// Supply valid payload, no Marble
	payload, err := json.Marshal(conf)
	require.NoError(err)
	mounter = assertionMounter{assert: assert, config: conf, usedTargets: make(map[string]bool), remountAsHostFS: false}
	assert.NoError(PreMain(string(payload), &mounter, fs))

	// Supply invalid payload, should fail
	payload = []byte("blablarubbish")
	assert.Error(PreMain(string(payload), &mounter, fs))
}

func TestPerformMounts(t *testing.T) {
	assert := assert.New(t)

	fs := afero.NewMemMapFs()

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

	// Same as above, but with remounting the hostfs as root
	confWithRemount := &config.Config{
		Exe:             "helloworld",
		Key:             "privatekey",
		Debug:           true,
		HeapSize:        512, //[MB]
		ProductID:       1,
		SecurityVersion: 1,
		Mounts:          []config.FileSystemMount{{Source: "/", Target: "/", Type: "hostfs", ReadOnly: false}, {Source: "/", Target: "/memfs", Type: "memfs", ReadOnly: true}},
	}

	mounter := assertionMounter{assert: assert, config: conf, usedTargets: make(map[string]bool), remountAsHostFS: false}
	assert.NoError(performUserMounts(*conf, &mounter, fs))

	conf.Mounts = []config.FileSystemMount{{Source: "/home/benjaminfranklin", Target: "/data", Type: "rubbishfs", ReadOnly: true}}
	assert.Error(performUserMounts(*conf, &mounter, fs))

	// Test '/' as host fs special case. Should work without an error, but we do not recommend doing this
	mounter = assertionMounter{assert: assert, config: confWithRemount, usedTargets: make(map[string]bool), remountAsHostFS: true}
	assert.NoError(performUserMounts(*confWithRemount, &mounter, fs))
}

func (a *assertionMounter) Mount(source string, target string, filesystem string, flags uintptr, data string) error {
	// Skip special mount calls for unit test, as we cannot check them against the configuration
	if target == "/" {
		return nil
	}
	if target == "/edg/mnt" {
		a.assert.EqualValues(mountTypeMemFS, filesystem)
		return nil
	}

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
	// Skip this check if we remount the host fs, as this will cause additional unspecified mount operations
	if currentMountPoint.Type == "" {
		if !a.remountAsHostFS {
			return errors.New("could not find equal mount declaration in supplied config")
		}
	}

	if filesystem == mountTypeHostFS {
		a.assert.EqualValues(currentMountPoint.Source, source)
		a.assert.EqualValues(currentMountPoint.Target, target)
		a.assert.EqualValues("hostfs", currentMountPoint.Type)
	} else if filesystem == mountTypeMemFS {
		if !a.remountAsHostFS {
			a.assert.EqualValues(memfsMountSourceDirectory+currentMountPoint.Target, source)
		} else {
			a.assert.EqualValues(currentMountPoint.Target, source)
		}
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

func (a *assertionMounter) Unmount(target string, flags int) error {
	// We only unmount the root memfs
	// Everything else what calls Unmount should fail
	a.assert.Equal("/", target)
	a.assert.Equal(syscall.MNT_FORCE, flags)

	return nil
}

func TestAddEnvVars(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Restore current env vars on exit
	defer restoreExistingEnvVars(os.Environ())

	// Get existing PWD env var from host system
	existingPwdEnvVar := os.Getenv("PWD")
	require.NotEmpty(os.Getenv("PWD"))

	// Set some existing env var which should vanish
	os.Setenv("EGO_INTEGRATION_TEST_PLS_FAIL_IF_I_EXIST", "bad")
	os.Setenv("EDG_WILL_I_SURVIVE?", "hopefully")

	//sane default values
	conf := &config.Config{
		Exe:             "helloworld",
		Key:             "privatekey",
		Debug:           true,
		HeapSize:        512, //[MB]
		ProductID:       1,
		SecurityVersion: 1,
		Env:             []config.EnvVar{{Name: "HELLO_WORLD", Value: "2"}, {Name: "PWD", Value: "/tmp/somedir", FromHost: true}, {Name: "NOT_EXISTING_ON_HOST", FromHost: true}, {Name: "NOT_EXISTING_ON_HOST_BUT_INITIALIZED", Value: "42", FromHost: true}},
	}

	// Apply env vars
	assert.NoError(addEnvVars(*conf))

	// Check if HELLO_WORLD was set correctly
	assert.Equal("2", os.Getenv("HELLO_WORLD"))

	// Check if PWD was taken correctly from host
	assert.Equal(existingPwdEnvVar, os.Getenv("PWD"))

	// Check if some other env var disappeared after applying
	envValue, envExists := os.LookupEnv("EGO_INTEGRATION_TEST_PLS_FAIL_IF_I_EXIST")
	assert.Empty(envValue)
	assert.False(envExists)

	// Check if EDG_ variables are preserved
	assert.Equal("hopefully", os.Getenv("EDG_WILL_I_SURVIVE?"))

	// Check if not existing fromHost variable is initialized empty if not existing
	envValue, envExists = os.LookupEnv("NOT_EXISTING_ON_HOST")
	assert.Empty(envValue)
	assert.False(envExists)

	// Check if not existing fromHost variable is initalized with the given tag if existing
	assert.Equal("42", os.Getenv("NOT_EXISTING_ON_HOST_BUT_INITIALIZED"))
}

func restoreExistingEnvVars(environ []string) {
	// Clear test environment
	os.Clearenv()

	// Restore all previously existing environment variables
	for _, envVar := range environ {
		splitString := strings.Split(envVar, "=")
		os.Setenv(splitString[0], splitString[1])
	}
}
