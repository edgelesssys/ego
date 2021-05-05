// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"os/exec"
)

type runner struct {
	run []*exec.Cmd
}

func (r *runner) Run(cmd *exec.Cmd) error {
	r.run = append(r.run, cmd)
	return nil
}

func (*runner) Output(cmd *exec.Cmd) ([]byte, error) {
	panic(cmd.Path)
}

func (*runner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	panic(cmd.Path)
}

func (*runner) ExitCode(cmd *exec.Cmd) int {
	return 2
}
