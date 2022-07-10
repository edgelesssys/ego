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

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <executable> [args...]",
		Short: "Run a signed executable in standalone mode",
		Long: `Run a signed executable in an enclave. You can pass arbitrary arguments to the enclave.

Environment variables are only readable from within the enclave if they start with ` + "`EDG_`" + `.

Set OE_SIMULATION=1 to run in simulation mode.`,
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagParsing:    true,
		DisableFlagsInUseLine: true,

		Run: func(cmd *cobra.Command, args []string) {
			exitCode, err := newCli().Run(args[0], args[1:])
			handleErr(err)
			os.Exit(exitCode)
		},
	}

	hideHelpFlag(cmd)
	return cmd
}
