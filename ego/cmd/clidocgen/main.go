// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Clidocgen generates a Markdown page describing all CLI commands.
package main

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/edgelesssys/ego/ego/ego/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var seeAlsoRegexp = regexp.MustCompile(`(?s)### SEE ALSO\n.+?\n\n`)

func main() {
	cobra.EnableCommandSorting = false
	rootCmd := cmd.NewRootCmd()
	rootCmd.DisableAutoGenTag = true

	// Generate Markdown for all commands.
	cmdList := &bytes.Buffer{}
	body := &bytes.Buffer{}
	for _, c := range rootCmd.Commands() {
		name := c.Name()
		fmt.Fprintf(cmdList, "* [%v](#ego-%v): %v\n", name, name, c.Short)
		if err := doc.GenMarkdown(c, body); err != nil {
			panic(err)
		}
	}

	// Remove "see also" sections. They list parent and child commands, which is not interesting for us.
	cleanedBody := seeAlsoRegexp.ReplaceAll(body.Bytes(), nil)

	fmt.Printf("Commands:\n\n%s\n%s", cmdList, cleanedBody)
}
