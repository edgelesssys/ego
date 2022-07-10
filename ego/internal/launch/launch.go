// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package launch

import (
	"io"
	"os"
	"os/exec"
)

// run launches an application with a CappedBuffer and also translates potential Edgeless RT / Open Enclave errors into more user-friendly ones.
func run(runner Runner, cmd *exec.Cmd) (int, error) {
	// force line buffering for stdout
	// otherwise it will be fully buffered because our stdout is not a tty
	path, err := exec.LookPath("stdbuf")
	if err != nil {
		return 1, err
	}
	cmd.Path = path
	cmd.Args = append([]string{"stdbuf", "-oL"}, cmd.Args...)

	cmd.Stdin = os.Stdin

	// capture stdout and stderr
	var stdout, stderr cappedBuffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	if err := runner.Run(cmd); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return 1, err
		}
	}
	return runner.ExitCode(cmd), findCommonError(string(stdout) + string(stderr))
}

// RunEnclave launches an user EGo enclave.
func RunEnclave(filename string, args []string, egoHostFilename string, egoEnclaveFilename string, runner Runner) (int, error) {
	enclaves := egoEnclaveFilename + ":" + filename
	args = append([]string{enclaves}, args...)
	cmd := exec.Command(egoHostFilename, args...)

	return run(runner, cmd)
}

// RunEnclaveMarblerun launches an user EGo enclave in MarbleRun mode, calling into MarbleRun's premain before launching user code.
func RunEnclaveMarblerun(filename string, egoHostFilename string, egoEnclaveFilename string, runner Runner) (int, error) {
	enclaves := egoEnclaveFilename + ":" + filename
	cmd := exec.Command(egoHostFilename, enclaves)

	// Enable the MarbleRun premain.
	if err := os.Setenv("EDG_EGO_PREMAIN", "1"); err != nil {
		return 1, err
	}

	return run(runner, cmd)
}

// Runner runs Cmd objects.
type Runner interface {
	Run(cmd *exec.Cmd) error
	Output(cmd *exec.Cmd) ([]byte, error)
	CombinedOutput(cmd *exec.Cmd) ([]byte, error)
	ExitCode(cmd *exec.Cmd) int
}

// OsRunner wraps Cmd objects from the real host system, not an unit test environment.
type OsRunner struct{}

// Run for OsRunner redirects to cmd.Run()
func (OsRunner) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

// Output for OsRunner redirects to cmd.Output()
func (OsRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

// CombinedOutput for OsRunner redirects to cmd.CombinedOutput()
func (OsRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

// ExitCode for OsRunner redirects to cmd.ProcessState.ExitCode()
func (OsRunner) ExitCode(cmd *exec.Cmd) int {
	return cmd.ProcessState.ExitCode()
}
