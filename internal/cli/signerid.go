// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

type eradump struct {
	SecurityVersion int
	ProductID       int
	UniqueID        string
	SignerID        string
}

func (c *Cli) signeridByKey(path string) {
	out, err := c.runner.Output(exec.Command("ego-oesign", "signerid", "-k", path))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

func (c *Cli) readEradumpJSONtoStruct(path string) *eradump {
	data, err := c.runner.Output(exec.Command("ego-oesign", "eradump", "-e", path))

	if err != nil {
		panic(err)
	}

	var dump eradump
	if err := json.Unmarshal(data, &dump); err != nil {
		panic(err)
	}
	return &dump
}

func (c *Cli) signeridByExecutable(path string) {
	dump := c.readEradumpJSONtoStruct(path)
	fmt.Println(dump.SignerID)
}

// Uniqueid prints the UniqueID of a signed executable.
func (c *Cli) Uniqueid(path string) {
	dump := c.readEradumpJSONtoStruct(path)
	fmt.Println(dump.UniqueID)
}

// Signerid prints the SignerID of a signed executable.
func (c *Cli) Signerid(path string) {
	if filepath.Ext(path) == ".pem" {
		c.signeridByKey(path)
	} else {
		c.signeridByExecutable(path)
	}
}
