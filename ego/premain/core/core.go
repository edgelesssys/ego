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
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"ego/config"

	"github.com/edgelesssys/marblerun/marble/premain"
	"github.com/spf13/afero"
)

// marblerunHostFSDirectory contains the expected path for the automatic host filesystem mount for Marbles
const marblerunHostFSDirectory = "/edg/hostfs"

// marblerunEnvVarFlag contains the name of the environment variable which holds the flag if we run a Marble
const marblerunEnvVarFlag = "EDG_EGO_PREMAIN"

// memfsDefaultMountDirectory contains the path where the default memfs should be mounted
const memfsDefaultMountDirectory = "/"

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
func PreMain(payload string, mounter Mounter, fs afero.Fs, hostEnviron []string) error {
	// Convert host environment string array to key-value map
	hostEnvironMap := make(map[string]string, len(hostEnviron))
	for _, envVar := range hostEnviron {
		splitString := strings.Split(envVar, "=")
		hostEnvironMap[splitString[0]] = splitString[1]
	}

	// Check if we run as a Marble or a normal EGo application
	isMarble := hostEnvironMap[marblerunEnvVarFlag] == "1"

	// Perform predefined mounts
	if err := performPredefinedMounts(mounter, isMarble); err != nil {
		return err
	}

	var config config.Config
	if len(payload) > 0 {
		// Load config from embedded payload
		if err := json.Unmarshal([]byte(payload), &config); err != nil {
			return err
		}

		// Perform user mounts based on embedded config
		if err := performUserMounts(config, mounter, fs); err != nil {
			return err
		}

		// Write embedded files to file system
		if err := writeFiles(config.Files, fs); err != nil {
			return err
		}
	}

	// Extract new environment variables
	if err := addEnvVars(config, hostEnvironMap); err != nil {
		return err
	}

	// If program is running as a Marble, continue with MarbleRun Premain.
	if isMarble {
		return premain.PreMainEgo()
	}

	return nil
}

func performPredefinedMounts(mounter Mounter, isMarble bool) error {
	// Mount host filesystem by default if running as a Marble
	if isMarble {
		if err := mounter.Mount("/", marblerunHostFSDirectory, mountTypeHostFS, 0, ""); err != nil {
			return err
		}
	}

	// Mount memory filesystem by default as root path
	if err := mounter.Mount("/", memfsDefaultMountDirectory, mountTypeMemFS, 0, ""); err != nil {
		return err
	}

	return nil
}

func performUserMounts(config config.Config, mounter Mounter, fs afero.Fs) error {
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

			if err := fs.MkdirAll(memfsMountSourceFull, 0o777); err != nil {
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

func writeFiles(files []config.File, fs afero.Fs) error {
	afs := afero.Afero{Fs: fs}

	for _, file := range files {
		buf, err := file.GetContent()
		if err != nil {
			return err
		}
		if err := afs.MkdirAll(filepath.Dir(file.Target), 0o700); err != nil {
			return err
		}
		if err := afs.WriteFile(file.Target, buf, 0o600); err != nil {
			return err
		}
	}

	return nil
}

func addEnvVars(config config.Config, hostEnvironMap map[string]string) error {
	// Copy all environment variables from host, and start from scratch
	newEnvVars := make(map[string]string)

	// Copy all special EDG_ environment variables
	for name, value := range hostEnvironMap {
		if strings.HasPrefix(name, "EDG_") {
			newEnvVars[name] = value
		}
	}

	// Copy all environment variable definitions from the enclave config
	for _, configEnvVar := range config.Env {
		// Only create new environment variable if value is not empty
		if configEnvVar.Value != "" {
			newEnvVars[configEnvVar.Name] = configEnvVar.Value
		}

		// Check if we can copy the env var from host
		if envVarFromHost, ok := hostEnvironMap[configEnvVar.Name]; configEnvVar.FromHost && ok {
			newEnvVars[configEnvVar.Name] = envVarFromHost
		}
	}

	// Now that we gathered the environment variables we want to keep or set, reset the environment and set all values
	for key, value := range newEnvVars {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return nil
}
