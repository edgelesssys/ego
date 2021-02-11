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

// Env runs a command in the EGo environment.
func (c *Cli) Env(filename string, args []string) {
	if filename == "go" {
		// "ego env go" should resolve to our Go compiler
		filename = filepath.Join(c.egoPath, "go", "bin", "go")
	}
	cmd := exec.Command(filename, args...)
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=1",
		"PATH="+filepath.Join(c.egoPath, "go", "bin")+":"+os.Getenv("PATH"),
		"GOROOT="+filepath.Join(c.egoPath, "go"))
	c.runAndExit(cmd)
}
