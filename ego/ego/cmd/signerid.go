// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSigneridCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "signerid <executable | key.pem>",
		Short:                 "Print the SignerID of a signed executable",
		Long:                  "Print the SignerID either from a signed executable or by reading a key file.",
		Args:                  cobra.ExactArgs(1),
		SilenceErrors:         true,
		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := newCli().Signerid(args[0])
			handleErr(err)
			if err != nil {
				return err
			}
			fmt.Println(id)
			return nil
		},
	}
}
