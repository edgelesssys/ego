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

func newUniqueidCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "uniqueid <executable>",
		Short:                 "Print the UniqueID of a signed executable",
		Long:                  "Print the UniqueID of a signed executable.",
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := newCli().Uniqueid(args[0])
			if err != nil {
				return err
			}
			fmt.Println(id)
			return nil
		},
	}
}
