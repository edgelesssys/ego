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
	"syscall"

	"github.com/edgelesssys/ego/internal/config"
	"github.com/edgelesssys/marblerun/marble/premain"
)

// Mounter defines an interface to use to mount the filesystem (usually syscall, mainly differs for unit tests)
type Mounter interface {
	Mount(source string, target string, filesystem string, flags uintptr, data string) error
}

// PreMain runs before the App's actual main routine and initializes the EGo enclave.
func PreMain(payload string, mounter Mounter) error {
	if len(payload) > 0 {
		// Load config from embedded payload
		var config config.Config
		if err := json.Unmarshal([]byte(payload), &config); err != nil {
			return err
		}

		// Perform mounts based on embedded config
		if err := performMounts(config, mounter); err != nil {
			return err
		}
	}

	// If program is running as a Marble, continue with Marblerun Premain.
	if os.Getenv("EDG_EGO_PREMAIN") == "1" {
		return premain.PreMain()
	}

	return nil
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
