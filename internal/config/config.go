// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"fmt"
)

// Config defines the structure of enclave.json, containing the settings for the enclave runtime
type Config struct {
	Exe             string            `json:"exe"`
	Key             string            `json:"key"`
	Debug           bool              `json:"debug"`
	HeapSize        int               `json:"heapSize"`
	ProductID       int               `json:"productID"`
	SecurityVersion int               `json:"securityVersion"`
	Mounts          []FileSystemMount `json:"mounts"`
}

// FileSystemMount defines a single mount point for the enclave's filesystem
// either from the enclave's host system (hostfs), or a virtual file system running in the enclave's memory (memfs).
type FileSystemMount struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Type     string `json:"type"`
	ReadOnly bool   `json:"readOnly"`
}

// Validate Exe, Key, HeapSize and Mounts
func (c *Config) Validate() error {
	if c.HeapSize == 0 {
		return fmt.Errorf("heapSize not set in config file")
	}
	if c.Exe == "" {
		return fmt.Errorf("exe not set in config file")
	}
	if c.Key == "" {
		return fmt.Errorf("key not set in config file")
	}

	// Validate file system mounts
	alreadyUsedMountPoints := make(map[string]bool, len(c.Mounts))
	for _, mountPoint := range c.Mounts {
		// Check if a target is defined multiple times. This will cause the syscall in the premain to return an error.
		if _, ok := alreadyUsedMountPoints[mountPoint.Target]; ok {
			fmt.Printf("ERROR: '%s': Mount point was defined multiple times. Check your configuration.", mountPoint.Target)
			return fmt.Errorf("mount target '%s' was defined multiple times", mountPoint.Target)
		}

		// Check if 'hostfs' or 'memfs' was set as type
		if mountPoint.Type != "hostfs" && mountPoint.Type != "memfs" {
			fmt.Printf("ERROR: '%s': Only mount types 'hostfs' and 'memfs' are accepted.", mountPoint.Target)
			return fmt.Errorf("an invalid mount type was specified: %s", mountPoint.Type)
		}

		// Warn user that 'memfs' source does nothing
		if mountPoint.Type == "memfs" && mountPoint.Source != "" && mountPoint.Source != "/" {
			fmt.Printf("WARNING: '%s': The mount point of type 'memfs' specified a source directory, will be ignored.", mountPoint.Target)
		}

		// Warn user that a read-only 'memfs' is useless
		if mountPoint.Type == "memfs" && mountPoint.ReadOnly {
			fmt.Printf("WARNING: '%s': The mount point of type 'memfs' is set as read-only, making it effectively useless. Check your configuration.", mountPoint.Target)
		}

		// Add already existing target to map of used targets for redefiniton checks
		alreadyUsedMountPoints[mountPoint.Target] = true
	}

	return nil
}
