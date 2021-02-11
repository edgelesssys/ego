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

	"github.com/spf13/afero"
)

// Runner runs Cmd objects.
type Runner interface {
	Run(cmd *exec.Cmd) error
	Output(cmd *exec.Cmd) ([]byte, error)
	CombinedOutput(cmd *exec.Cmd) ([]byte, error)
	ExitCode(cmd *exec.Cmd) int
}

// Cli implements the ego commands.
type Cli struct {
	runner  Runner
	fs      afero.Afero
	egoPath string
}

// NewCli creates a new Cli object.
func NewCli(runner Runner, fs afero.Fs) *Cli {
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

func (c *Cli) run(cmd *exec.Cmd) int {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := c.runner.Run(cmd); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			panic(err)
		}
	}
	return c.runner.ExitCode(cmd)
}
