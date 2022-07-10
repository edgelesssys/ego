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

func newMarblerunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "marblerun <executable>",
		Short: "Run a signed executable as a MarbleRun Marble",
		Long: `Run a signed executable as a MarbleRun Marble.
Requires a running MarbleRun Coordinator instance.
Environment variables are only readable from within the enclave if they start with "EDG_" and
will be extended/overwritten with the ones specified in the manifest.

Requires the following configuration environment variables:
  EDG_MARBLE_COORDINATOR_ADDR   The Coordinator address
  EDG_MARBLE_TYPE               The type of this Marble (as specified in the manifest)
  EDG_MARBLE_DNS_NAMES          The alternative DNS names for this Marble's TLS certificate
  EDG_MARBLE_UUID_FILE          The location where this Marble will store its UUID

Set OE_SIMULATION=1 to run in simulation mode.`,
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,

		Run: func(cmd *cobra.Command, args []string) {
			exitCode, err := newCli().Marblerun(args[0])
			handleErr(err)
			os.Exit(exitCode)
		},
	}
}
