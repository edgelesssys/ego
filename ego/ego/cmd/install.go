// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install [component]",
		Short: "Install drivers and other components",
		Long: `Install drivers and other components. The components that you can install depend on your operating system and its version.
Use "ego install" to list the available components for your system.`,
		Args:                  cobra.MaximumNArgs(1),
		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			var component string
			if len(args) == 1 {
				component = args[0]
			}
			return newCli().Install(askInstall, component)
		},
	}
}

// Asks the user whether he wants to execute the commands in listOfActions and returns his choice
func askInstall(listOfActions string) bool {
	fmt.Println("The following commands will be executed:\n\n" + listOfActions + "\n")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Continue installation? [y/n]: ")
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
