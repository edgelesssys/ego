// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"log/slog"
	"os"

	"github.com/edgelesssys/ego/ego/ego/cmd"
	"github.com/edgelesssys/ego/ego/internal/config"
)

// Don't touch! Automatically injected at build-time.
var (
	version   = "0.0.0"
	gitCommit = "0000000000000000000000000000000000000000"
)

func main() {
	if config.LogJSON {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}
	slog.Info("EGo", "version", version, "git_commit", gitCommit)
	if cmd.Execute() != nil {
		os.Exit(1)
	}
}
