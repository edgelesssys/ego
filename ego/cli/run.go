// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"path/filepath"
	"runtime/debug"
	"slices"

	"github.com/edgelesssys/ego/ego/internal/launch"
)

// Run runs a signed executable in standalone mode.
func (c *Cli) Run(filename string, args []string) (int, error) {
	buildInfo, err := c.checkUnsupportedImports(filename)
	if err != nil {
		return 1, err
	}
	return launch.RunEnclave(filename, args, c.getEgoHostPath(), c.getEgoEnclavePath(buildInfo), c.runner)
}

// Marblerun runs a signed executable as a MarbleRun Marble.
func (c *Cli) Marblerun(filename string) (int, error) {
	buildInfo, err := c.checkUnsupportedImports(filename)
	if err != nil {
		return 1, err
	}
	return launch.RunEnclaveMarblerun(filename, c.getEgoHostPath(), c.getEgoEnclavePath(buildInfo), c.runner)
}

func (c *Cli) getEgoHostPath() string {
	return filepath.Join(c.egoPath, "bin", "ego-host")
}

func (c *Cli) getEgoEnclavePath(buildInfo *debug.BuildInfo) string {
	egoEnclave := "ego-enclave"
	if slices.ContainsFunc(buildInfo.Settings, func(bs debug.BuildSetting) bool { return bs.Key == "GOFIPS140" }) {
		// if app is built with FIPS enabled, also use the FIPS-enabled EGo runtime
		egoEnclave = "ego-enclave-fips140"
	}
	return filepath.Join(c.egoPath, "share", egoEnclave)
}
