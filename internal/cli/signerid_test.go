// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestUniqueid(t *testing.T) {
	assert := assert.New(t)

	cli := NewCli(signeridRunner{}, afero.NewMemMapFs())

	res, err := cli.Uniqueid("foo")
	assert.NoError(err)
	assert.Equal("uid foo", res)
}

func TestSignerid(t *testing.T) {
	assert := assert.New(t)

	cli := NewCli(signeridRunner{}, afero.NewMemMapFs())

	res, err := cli.Signerid("foo")
	assert.NoError(err)
	assert.Equal("sid foo", res)

	res, err = cli.Signerid("foo.pem")
	assert.NoError(err)
	assert.Equal("id foo.pem", res)
}

type signeridRunner struct{}

func (signeridRunner) Run(cmd *exec.Cmd) error {
	panic(cmd.Path)
}

func (signeridRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	if filepath.Base(cmd.Path) != "ego-oesign" || len(cmd.Args) != 4 {
		return nil, errors.New("unexpected cmd")
	}
	if cmd.Args[1] == "signerid" && cmd.Args[2] == "-k" {
		return []byte("id " + cmd.Args[3]), nil
	}
	if cmd.Args[1] == "eradump" && cmd.Args[2] == "-e" {
		path := cmd.Args[3]
		return []byte(`{"uniqueid":"uid ` + path + `","signerid":"sid ` + path + `"}`), nil
	}
	return nil, errors.New("unexpected subcommand")
}

func (signeridRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	panic(cmd.Path)
}

func (signeridRunner) ExitCode(cmd *exec.Cmd) int {
	panic(cmd.Path)
}
