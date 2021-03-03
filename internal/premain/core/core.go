// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package core

import (
	"encoding/json"
	"os"
	"syscall"

	"github.com/edgelesssys/ego/internal/config"
	"github.com/edgelesssys/marblerun/marble/premain"
)

// PreMain runs before the App's actual main routine and initializes the EGo enclave.
func PreMain(payload string) error {
	if len(payload) > 0 {
		// Load config from embedded payload
		var config config.Config
		if err := json.Unmarshal([]byte(payload), &config); err != nil {
			return err
		}

		// Perform mounts
		for _, mountPoint := range config.Mounts {
			var err error

			switch mountPoint.Type {
			case "hostfs":
				if mountPoint.ReadOnly {
					err = syscall.Mount(mountPoint.Source, mountPoint.Target, "oe_host_file_system", syscall.MS_RDONLY, "")
				} else {
					err = syscall.Mount(mountPoint.Source, mountPoint.Target, "oe_host_file_system", 0, "")
				}

			case "memfs":
				if mountPoint.ReadOnly {
					err = syscall.Mount("/", mountPoint.Target, "edg_memfs", syscall.MS_RDONLY, "")
				} else {
					err = syscall.Mount("/", mountPoint.Target, "edg_memfs", 0, "")
				}
			}

			if err != nil {
				return err
			}
		}
	}

	// If program is running as a Marble, continue with Marblerun Premain.
	if os.Getenv("EDG_EGO_PREMAIN") == "1" {
		return premain.PreMain()
	}

	return nil
}
