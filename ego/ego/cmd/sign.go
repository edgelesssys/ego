// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"github.com/spf13/cobra"
)

func newSignCmd() *cobra.Command {
	return &cobra.Command{
		Use:           "sign [executable | config.json]",
		Short:         "Sign an executable built with ego-go",
		Long:          "Sign an executable built with ego-go. Executables must be signed before they can be run in an enclave.",
		SilenceErrors: true,
		Example: `  ego sign <executable>
    Generates a new key "private.pem" and a default configuration "enclave.json" in the current directory and signs the executable.

  ego sign
    Searches in the current directory for "enclave.json" and signs the therein provided executable.

  ego sign <config.json>
    Signs an executable according to a given configuration.`,
		Args:                  cobra.MaximumNArgs(1),
		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			var filename string
			if len(args) > 0 {
				filename = args[0]
			}
			err := newCli().Sign(filename)
			handleErr(err)
			return err // nil if no error
		},
	}
}
