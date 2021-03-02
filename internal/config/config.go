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

// Validate Exe, Key, HeapSize
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
	return nil
}
