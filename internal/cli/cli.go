// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func (c *Cli) run(cmd *exec.Cmd) (int, error) {
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

	if err := c.runner.Run(cmd); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return 1, err
		}
	}
	return c.runner.ExitCode(cmd), findCommonError(string(stdout) + string(stderr))
}

func (c *Cli) getOesignPath() string {
	return filepath.Join(c.egoPath, "bin", "ego-oesign")
}

type cappedBuffer []byte

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if len(*b) < 1000 {
		*b = append(*b, p...)
	}
	return len(p), nil
}

func findCommonError(s string) error {
	switch {
	case strings.Contains(s, "enclave_initialize failed (err=0x1001)"):
		return ErrEnclIniFail
	case strings.Contains(s, "oe_sgx_is_valid_attributes failed: attributes = 0"):
		return ErrValidAttr0
	case strings.Contains(s, "ELF image is not a PIE or shared object"):
		return ErrElfNoPie
	case strings.Contains(s, "Failed to open Intel SGX device"):
		return ErrSGXOpenFail
	default:
		return nil
	}
}

// ErrOEInvalidImg is a representation of Open Enclaves OE_INVALID_IMAGE return code.
var ErrOEInvalidImg = errors.New("OE_INVALID_IMAGE")

// ErrOEFailure is a representation of Open Enclaves OE_FAILURE return code.
var ErrOEFailure = errors.New("OE_FAILURE")

// ErrOEPlatform is a representation of Open Enclaves OE_PLATFORM_ERROR return code.
var ErrOEPlatform = errors.New("OE_PLATFORM_ERROR")

// ErrExtUnknown is a unknown error from an external tool.
var ErrExtUnknown = errors.New("unknown external error")

// ErrEnclIniFail is an Open Enclave error where enclave_initialize fails with error code 0x1001.
// This likely occures if the signature of the binary is invalid and the binary needs to be resigned.
var ErrEnclIniFail = fmt.Errorf("%w: enclave_initialize failed (err=0x1001)", ErrOEPlatform)

// ErrValidAttr0 is an Open Enclave error where oe_sgx_is_valid_attributes fails.
// This likely occures if an unsigned binary is run.
var ErrValidAttr0 = fmt.Errorf("%w: oe_sgx_is_valid_attributes failed: attributes = 0", ErrOEFailure)

// ErrElfNoPie is an Open Enclave error where the ELF image is not a PIE or shared object.
// This likely occures if a binary is run which was not built with ego-go.
var ErrElfNoPie = fmt.Errorf("%w: ELF image is not a PIE or shared object", ErrOEInvalidImg)

// ErrSGXOpenFail is an Open Enclave error where OE failes to open the Intel SGX device.
// This likely occures if a system does not support SGX or the required module is missing.
var ErrSGXOpenFail = fmt.Errorf("%w: Failed to open Intel SGX device", ErrOEPlatform)
