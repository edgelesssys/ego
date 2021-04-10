// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type eradump struct {
	SecurityVersion int
	ProductID       int
	UniqueID        string
	SignerID        string
}

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
		if strings.Contains(out, ErrOECrypto.Error()) {
			return "", ErrOECrypto
		}
	}

	return "", err
}

func (c *Cli) readEradumpJSONtoStruct(path string) (*eradump, error) {
	data, err := c.runner.Output(exec.Command(c.getOesignPath(), "eradump", "-e", path))

	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return nil, errors.New(string(err.Stderr))
		}
		return nil, err
	}

	var dump eradump
	if err := json.Unmarshal(data, &dump); err != nil {
		return nil, err
	}
	return &dump, nil
}

func (c *Cli) signeridByExecutable(path string) (string, error) {
	dump, err := c.readEradumpJSONtoStruct(path)
	if err != nil {
		return "", err
	}
	return dump.SignerID, nil
}

// Uniqueid returns the UniqueID of a signed executable.
func (c *Cli) Uniqueid(path string) (string, error) {
	dump, err := c.readEradumpJSONtoStruct(path)
	if err != nil {
		return "", err
	}
	return dump.UniqueID, nil
}

// Signerid returns the SignerID of a signed executable.
func (c *Cli) Signerid(path string) (string, error) {
	if filepath.Ext(path) == ".pem" {
		return c.signeridByKey(path)
	}
	return c.signeridByExecutable(path)
}
