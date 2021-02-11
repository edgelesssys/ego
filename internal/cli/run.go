// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Run runs a signed executable in standalone mode.
func (c *Cli) Run(filename string, args []string) {
	enclaves := filepath.Join(c.egoPath, "share", "ego-enclave") + ":" + filename
	args = append([]string{enclaves}, args...)
	os.Setenv("EDG_EGO_PREMAIN", "0")
	cmd := exec.Command("ego-host", args...)
	c.runAndExit(cmd)
}

// Marblerun runs a signed executable as a Marblerun Marble.
func (c *Cli) Marblerun(filename string) {
	enclaves := filepath.Join(c.egoPath, "share", "ego-enclave") + ":" + filename
	os.Setenv("EDG_EGO_PREMAIN", "1")
	cmd := exec.Command("ego-host", enclaves)
	c.runAndExit(cmd)
}
