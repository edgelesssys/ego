// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

// Install ...
func (c *Cli) Install(ask func(string) bool, component string) error {
	// TODO get relevant cpuid field(s) and pass them to install
	return c.install(ask, component, "https://raw.githubusercontent.com/edgelesssys/ego/feat/install/src/install.json")
}

func (c *Cli) install(ask func(string) bool, component string, jsonURL string) error {
	return nil
}
