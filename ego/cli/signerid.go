// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ego/internal/launch"
)

const (
	offsetSigstruct = 144
	offsetModulus   = 128
	offsetMRENCLAVE = 960
)

func (c *Cli) signeridByKey(path string) (string, error) {
	outBytes, err := c.runner.Output(exec.Command(c.getOesignPath(), "signerid", "-k", path))
	out := string(outBytes)
	if err == nil {
		return out, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		// oesign tends to print short errors to stderr and logs to stdout
		if len(err.Stderr) > 0 {
			return "", errors.New(string(err.Stderr))
		}
		fmt.Fprintln(os.Stderr, out)
		if strings.Contains(out, launch.ErrOECrypto.Error()) {
			return "", launch.ErrOECrypto
		}
	}

	return "", err
}

func (c *Cli) signeridByExecutable(path string) (string, error) {
	const modulusSize = 384
	modulus, err := c.readDataFromELF(path, oeinfoSectionName, offsetSigstruct+offsetModulus, modulusSize)
	if err != nil {
		return "", err
	}
	if bytes.Equal(modulus, make([]byte, modulusSize)) {
		// if enclave is unsigned, return all zeros
		return strings.Repeat("0", 64), nil
	}

	// MRSIGNER is the sha256 of the modulus
	sum := sha256.Sum256(modulus)
	return hex.EncodeToString(sum[:]), nil
}

// Uniqueid returns the UniqueID of a signed executable.
func (c *Cli) Uniqueid(path string) (string, error) {
	data, err := c.readDataFromELF(path, oeinfoSectionName, offsetSigstruct+offsetMRENCLAVE, 32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

// Signerid returns the SignerID of a signed executable.
func (c *Cli) Signerid(path string) (string, error) {
	if filepath.Ext(path) == ".pem" {
		return c.signeridByKey(path)
	}
	return c.signeridByExecutable(path)
}
