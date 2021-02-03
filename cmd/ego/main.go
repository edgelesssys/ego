// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/edgelesssys/ego/internal/cli"
	"github.com/spf13/afero"
)

func main() {
	if len(os.Args) < 2 {
		help("")
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]
	cli := cli.NewCli(runner{}, afero.NewOsFs())

	switch cmd {
	case "sign":
		var filename string
		if len(args) > 0 {
			filename = args[0]
		}
		err := cli.Sign(filename)
		if err == nil {
			return
		}
		fmt.Println(err)
		if !os.IsNotExist(err) {
			// print error only
			return
		}
		// also print usage
		fmt.Println()
	case "run":
		if len(args) > 0 {
			os.Exit(cli.Run(args[0], args[1:]))
		}
	case "marblerun":
		if len(args) == 1 {
			os.Exit(cli.Marblerun(args[0]))
		}
	case "signerid":
		if len(args) == 1 {
			id, err := cli.Signerid(args[0])
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(id)
			return
		}
	case "uniqueid":
		if len(args) == 1 {
			id, err := cli.Uniqueid(args[0])
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(id)
			return
		}
	case "env":
		if len(args) > 0 {
			os.Exit(cli.Env(args[0], args[1:]))
		}
	case "help":
		if len(args) == 1 {
			help(args[0])
			return
		}
	}

	help(cmd)
}

func help(cmd string) {
	me := os.Args[0]
	var s string

	switch cmd {
	case "sign":
		s = `sign [executable | config.json]

Sign an executable built with ego-go. Executables must be signed before they can be run in an enclave.

This command can be used in different modes:

ego sign <executable>
  Generates a new key "private.pem" and a default configuration "enclave.json" in the current directory and signs the executable.

ego sign
  Searches in the current directory for "enclave.json" and signs the therein provided executable.

ego sign <config.json>
  Signs an executable according to a given configuration.`

	case "run":
		s = `run <executable> [args...]

Run a signed executable in an enclave. You can pass arbitrary arguments to the enclave.

Environment variables are only readable from within the enclave if they start with "EDG_".

Set OE_SIMULATION=1 to run in simulation mode.`

	case "marblerun":
		s = `marblerun <executable>

Run a signed executable as a Marblerun Marble.
Requires a running Marblerun Coordinator instance.
Environment variables are only readable from within the enclave if they start with "EDG_" and
will be extended/overwritten with the ones specified in the manifest.

Requires the following configuration environment variables:
  EDG_MARBLE_COORDINATOR_ADDR   The Coordinator address
  EDG_MARBLE_TYPE               The type of this Marble (as specified in the manifest)
  EDG_MARBLE_DNS_NAMES          The alternative DNS names for this Marble's TLS certificate
  EDG_MARBLE_UUID_FILE          The location where this Marble will store its UUID

Set OE_SIMULATION=1 to run in simulation mode.`

	case "env":
		s = `env ...

Run a command within the EGo environment. For example, run
` + me + ` env make
to build a Go project that uses a Makefile.`

	case "signerid":
		s = `signerid <executable | key.pem>

Print the SignerID either from a signed executable or by reading a keyfile.`

	case "uniqueid":
		s = `uniqueid <executable>

Print the UniqueID of a signed executable.`

	default:
		s = `<command> [arguments]

Commands:
  sign        Sign an executable built with ego-go.
  run         Run a signed executable in standalone mode.
  marblerun   Run a signed executable as a Marblerun Marble.
  signerid    Print the SignerID of a signed executable.
  uniqueid    Print the UniqueID of a signed executable.
  env         Run a command in the EGo environment.

Use "` + me + ` help <command>" for more information about a command.`
	}

	fmt.Println("Usage: " + me + " " + s)
}

type runner struct{}

func (runner) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (runner) Output(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

func (runner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

func (runner) ExitCode(cmd *exec.Cmd) int {
	return cmd.ProcessState.ExitCode()
}
