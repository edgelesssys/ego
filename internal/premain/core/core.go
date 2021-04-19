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
	"path"
	"sort"
	"strings"
	"syscall"

	"github.com/edgelesssys/ego/internal/config"
	"github.com/edgelesssys/marblerun/marble/premain"
	"github.com/spf13/afero"
)

// memfsMountSourceDirectory contains the 'global' memfs <-> 'mounted' memfs base directory
const memfsMountSourceDirectory = "/edg/mnt"

// mountTypeHostFS is the parameter for the mount filesystem type of the host filesystem in Open Enclave / Edgeless RT
const mountTypeHostFS = "oe_host_file_system"

// mountTypeMemfsFS is the paramter for the mount filesystem type of the in-memory filesystem in Edgeless RT
const mountTypeMemFS = "edg_memfs"

// Mounter defines an interface to use to mount the filesystem (usually syscall, mainly differs for unit tests)
type Mounter interface {
	Mount(source string, target string, filesystem string, flags uintptr, data string) error
	Unmount(target string, flags int) error
}

// PreMain runs before the App's actual main routine and initializes the EGo enclave.
func PreMain(payload string, mounter Mounter, fs afero.Fs) error {
	var config config.Config
	if len(payload) > 0 {
		// Load config from embedded payload
		if err := json.Unmarshal([]byte(payload), &config); err != nil {
			return err
		}

		// Perform mounts based on embedded config
		if err := performMounts(config, mounter, fs); err != nil {
			return err
		}
	}

	// Extract new environment variables
	if err := addEnvVars(config); err != nil {
		return err
	}

	// If program is running as a Marble, continue with Marblerun Premain.
	if os.Getenv("EDG_EGO_PREMAIN") == "1" {
		return premain.PreMain()
	}

	return nil
}

func performMounts(config config.Config, mounter Mounter, fs afero.Fs) error {
	// Check if / is mounted (handled in enc.cpp, should be memfs)
	_, err := os.Stat("/")
	if err != nil {
		return errors.New("no root filesystem mounted")
	}

	// Sort slice by length of target, so that we can catch "/" as special case without having to double loop or build another data structure
	sort.Slice(config.Mounts, func(i, j int) bool {
		return len(config.Mounts[i].Target) < len(config.Mounts[j].Target)
	})

	// Prepare the mounts specified in the config
	var rootIsHostFS bool
	for _, mountPoint := range config.Mounts {
		// Oh oh, special case!
		// Check if user specified to remount root as hostfs (possibly insecure)
		// Unmount premounted root if "/" was specified as target
		if !rootIsHostFS && mountPoint.Target == "/" && mountPoint.Type == "hostfs" {
			if err := mounter.Unmount("/", syscall.MNT_FORCE); err != nil {
				return err
			}
			fmt.Println("WARNING: Remounted '/' to hostfs. This might be insecure. Please only use this for testing purposes.")
			// Remounting memfs to /edg/mnt, we just keep it as base for memfs mounts
			if err := mounter.Mount("/", "/", "oe_host_file_system", 0, ""); err != nil {
				return err
			}
			if err := mounter.Mount("/", "/edg/mnt", "edg_memfs", 0, ""); err != nil {
				return err
			}

			rootIsHostFS = true
			continue
		}

		// Setup flags for read-only or read-write
		var flags uintptr
		if mountPoint.ReadOnly {
			flags = syscall.MS_RDONLY
		}

		// Select either hostfs (oe_host_file_system) and memfs (edg_memfs)
		var filesystem string
		switch mountPoint.Type {
		case "hostfs":
			filesystem = mountTypeHostFS
		case "memfs":
			filesystem = mountTypeMemFS

			memfsMountSourceFull := path.Join(memfsMountSourceDirectory, mountPoint.Target)

			if err := fs.MkdirAll(memfsMountSourceFull, 0777); err != nil {
				return err
			}

			// BEWARE: Confusion lies ahead.
			// The source is expected to be the *relative* path from the memfs root
			// Open Enclave / Edgeless RT does not fully resolve the path based on previous mounts
			// Thus, we either have / (if not remouted) or /edg/memfs (if remouted) as base for the memfs.
			// Depending on that, we need to shorten the source path for the latter case. Otherwise, use the full one.

			if rootIsHostFS {
				mountPoint.Source = mountPoint.Target
			} else {
				mountPoint.Source = memfsMountSourceFull
			}
		// This should not happen, as 'ego sign' is supposed to validate the config before embedding & signing it
		default:
			return errors.New("encountered an unknown filesystem type in configuration")
		}

		// Perform the mount
		if err := mounter.Mount(mountPoint.Source, mountPoint.Target, filesystem, flags, ""); err != nil {
			return err
		}
	}

	return nil
}

func addEnvVars(config config.Config) error {
	// Copy all environment variables from host, and start from scratch
	existingEnvVars := os.Environ()
	newEnvVars := make(map[string]string)

	// Set OE_IS_ENCLAVE to 1
	newEnvVars["OE_IS_ENCLAVE"] = "1"

	// Copy all special EDG_ environment variables
	for _, envVar := range existingEnvVars {
		if strings.HasPrefix(envVar, "EDG_") {
			splitString := strings.Split(envVar, "=")
			newEnvVars[splitString[0]] = splitString[1]
		}
	}

	// Copy all environment variable definitions from the enclave config
	for _, configEnvVar := range config.Env {
		// Only create new environment variable if value is not empty
		if configEnvVar.Value != "" {
			newEnvVars[configEnvVar.Name] = configEnvVar.Value
		}

		// Check if we can copy the env var from host
		if envVarFromHost := os.Getenv(configEnvVar.Name); configEnvVar.FromHost && envVarFromHost != "" {
			newEnvVars[configEnvVar.Name] = envVarFromHost
		}
	}

	// Now that we gathered the environment variables we want to keep or set, reset the environment and set all values
	os.Clearenv()
	for key, value := range newEnvVars {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return nil
}
