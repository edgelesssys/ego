// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"github.com/spf13/cobra"
)

func newBundleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bundle <executable> [output]",
		Short: "Bundle a signed executable with the current EGo runtime into a single executable",
		Long: `Bundles a signed executable with the current EGo runtime into a single all-in-one executable.

This option is useful if you plan to build and run your enclave on separate systems,
which might not have EGo installed or uses a different version that is not compatible to the version the executable was built with.

Note that the SGX driver and libraries still need to be installed on the target system in order to execute the bundled executable without issues.

If no output filename is specified, the output binary will be created with the same name as the source executable, appended with "-bundle".`,
		Args:                  cobra.RangeArgs(1, 2),
		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			var outputFilename string
			if len(args) > 1 {
				outputFilename = args[1]
			}
			return newCli().Bundle(args[0], outputFilename)
		},
	}
}
