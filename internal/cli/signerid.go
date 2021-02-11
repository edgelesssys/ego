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

func signeridByKey(path string) {
	out, err := exec.Command("ego-oesign", "signerid", "-k", path).Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

func readEradumpJSONtoStruct(path string) *eradump {
	data, err := exec.Command("ego-oesign", "eradump", "-e", path).Output()

	if err != nil {
		panic(err)
	}

	var dump eradump
	if err := json.Unmarshal(data, &dump); err != nil {
		panic(err)
	}
	return &dump
}

func signeridByExecutable(path string) {
	dump := readEradumpJSONtoStruct(path)
	fmt.Println(dump.SignerID)
}

// Uniqueid prints the UniqueID of a signed executable.
func Uniqueid(path string) {
	dump := readEradumpJSONtoStruct(path)
	fmt.Println(dump.UniqueID)
}

// Signerid prints the SignerID of a signed executable.
func Signerid(path string) {
	if filepath.Ext(path) == ".pem" {
		signeridByKey(path)
	} else {
		signeridByExecutable(path)
	}
}
