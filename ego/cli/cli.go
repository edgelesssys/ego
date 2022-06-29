// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"os"
	"path/filepath"

	"ego/internal/launch"

	"github.com/spf13/afero"
)

// Cli implements the ego commands.
type Cli struct {
	runner  launch.Runner
	fs      afero.Afero
	egoPath string
}

// NewCli creates a new Cli object.
func NewCli(runner launch.Runner, fs afero.Fs) *Cli {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return &Cli{
		runner:  runner,
		fs:      afero.Afero{Fs: fs},
		egoPath: filepath.Dir(filepath.Dir(exe)),
	}
}

func (c *Cli) getOesignPath() string {
	return filepath.Join(c.egoPath, "bin", "ego-oesign")
}
