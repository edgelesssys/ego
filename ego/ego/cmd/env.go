// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "env ...",
		Short: "Run a command in the EGo environment",
		Long:  "Run a command within the EGo environment.",
		Example: `  ego env make
    Builds a Go project that uses a Makefile.`,
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,

		Run: func(cmd *cobra.Command, args []string) {
			os.Exit(newCli().Env(args[0], args[1:]))
		},
	}
}
