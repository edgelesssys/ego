// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvNoArgs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	runner := runner{}
	cli := NewCli(&runner, afero.NewMemMapFs())

	exitcode, _ := cli.Env("foo", nil)
	assert.Equal(2, exitcode)
	require.Len(runner.run, 1)
	cmd := runner.run[0]
	assert.Equal("foo", cmd.Path)
	assert.Len(cmd.Args, 1)
}

func TestEnvArgs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	runner := runner{}
	cli := NewCli(&runner, afero.NewMemMapFs())

	exitcode, _ := cli.Env("foo", []string{"arg1", "arg2"})
	assert.Equal(2, exitcode)
	require.Len(runner.run, 1)
	cmd := runner.run[0]
	assert.Equal("foo", cmd.Path)
	require.Len(cmd.Args, 3)
	assert.Equal(cmd.Args[1], "arg1")
	assert.Equal(cmd.Args[2], "arg2")
}

func TestEnvGo(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	runner := runner{}
	cli := NewCli(&runner, afero.NewMemMapFs())

	exitcode, _ := cli.Env("go", nil)
	assert.Equal(2, exitcode)
	require.Len(runner.run, 1)
	path := runner.run[0].Path
	assert.Equal("go", filepath.Base(path))
	assert.Equal("bin", filepath.Base(filepath.Dir(path)))
}
