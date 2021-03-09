// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package core

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"syscall"

	"github.com/edgelesssys/ego/internal/config"
	"github.com/edgelesssys/marblerun/marble/premain"
)

// Mounter defines an interface to use to mount the filesystem (usually syscall, mainly differs for unit tests)
type Mounter interface {
	Mount(source string, target string, filesystem string, flags uintptr, data string) error
}

// PreMain runs before the App's actual main routine and initializes the EGo enclave.
func PreMain(payload string, mounter Mounter) ([]string, error) {
	var config config.Config
	if len(payload) > 0 {
		// Load config from embedded payload
		if err := json.Unmarshal([]byte(payload), &config); err != nil {
			return nil, err
		}

		// Perform mounts based on embedded config
		if err := performMounts(config, mounter); err != nil {
			return nil, err
		}
	}

	// Extract new environment variables
	if err := addEnvVars(config); err != nil {
		return nil, err
	}

	// If program is running as a Marble, continue with Marblerun Premain.
	if os.Getenv("EDG_EGO_PREMAIN") == "1" {
		if err := premain.PreMain(); err != nil {
			return nil, err
		}
	}

	// Return new environment variables to the caller
	return os.Environ(), nil
}

func performMounts(config config.Config, mounter Mounter) error {
	for _, mountPoint := range config.Mounts {
		// Setup flags for read-only or read-write
		var flags uintptr
		if mountPoint.ReadOnly {
			flags = syscall.MS_RDONLY
		}

		// Mount filesystem
		var err error
		switch mountPoint.Type {
		case "hostfs":
			err = mounter.Mount(mountPoint.Source, mountPoint.Target, "oe_host_file_system", flags, "")
		case "memfs":
			err = mounter.Mount("/", mountPoint.Target, "edg_memfs", flags, "")
		// This should not happen, as 'ego sign' is supposed to validate the config before embedding & signing it
		default:
			return errors.New("encountered an unknown filesystem type in configuration")
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func addEnvVars(config config.Config) error {
	// Copy all environment variables from host, and start from scratch
	existingEnvVars := os.Environ()
	newEnvVars := make(map[string]string)

	// Copy all special EDG_ environment variables
	for _, envVar := range existingEnvVars {
		if strings.HasPrefix(envVar, "EDG_") {
			splitString := strings.Split(envVar, "=")
			newEnvVars[splitString[0]] = os.Getenv(splitString[0])
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
