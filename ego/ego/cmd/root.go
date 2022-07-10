// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import "github.com/spf13/cobra"

func Execute() error {
	cobra.EnableCommandSorting = false
	return NewRootCmd().Execute()
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              "ego",
		Short:            "Manage and run EGo enclaves",
		Long:             "Manage and run EGo enclaves.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) { cmd.SilenceUsage = true },
	}

	rootCmd.AddCommand(newSignCmd())
	rootCmd.AddCommand(newRunCmd())
	rootCmd.AddCommand(newMarblerunCmd())
	rootCmd.AddCommand(newBundleCmd())
	rootCmd.AddCommand(newSigneridCmd())
	rootCmd.AddCommand(newUniqueidCmd())
	rootCmd.AddCommand(newEnvCmd())
	rootCmd.AddCommand(newInstallCmd())

	return rootCmd
}
