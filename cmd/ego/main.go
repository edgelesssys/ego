// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"

	"github.com/edgelesssys/ego/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		help("")
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "sign":
		var filename string
		if len(args) > 0 {
			filename = args[0]
		}
		cli.Sign(filename)
		return
	case "signerid":
		if len(args) == 1 {
			cli.Signerid(args[0])
			return
		}
	case "uniqueid":
		if len(args) == 1 {
			cli.Uniqueid(args[0])
			return
		}
	case "run":
		if len(args) > 0 {
			cli.Run(args[0], args[1:])
			return
		}
	case "marblerun":
		if len(args) == 1 {
			cli.Marblerun(args[0])
			return
		}
	case "env":
		if len(args) > 0 {
			cli.Env(args[0], args[1:])
			return
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
		s = `sign [executable|json]

Sign an executable built with ego-go. Executables must be signed before they can be run in an enclave.
There are 3 different ways of signing an executable:

1. ego sign
Searches in the current directory for enclave.json and signs the therein provided executable.

2. ego sign <executable>
Searches in the current directory for enclave.json.
If no such file is found, create one with default parameters for the executable.
Use enclave.json to sign the executable.

3. ego sign <json>
Reads json and signs the therein provided executable.`

	case "run":
		s = `run <executable> [args...]

Run a signed executable in an enclave. You can pass arbitrary arguments to the enclave.

Environment variables are only readable from within the enclave if they start with "EDG_".

Set OE_SIMULATION=1 to run in simulation mode.`

	case "marblerun":
		s = `marblerun <executable>

Run a signed executable as a Marblerun marble.
Requires a running Marblerun Coordinator instance.
Environment variables are only readable from within the enclave if they start with "EDG_".
Environment variables will be extended/overwritten with the ones specified in the manifest.
Requires the following configuration environment variables:
		- EDG_MARBLE_COORDINATOR_ADDR: The Coordinator address
		- EDG_MARBLE_TYPE: The type of this marble (as specified in the manifest)
		- EDG_MARBLE_DNS_NAMES: The alternative DNS names for this marble's TLS certificate
		- EDG_MARBLE_UUID_FILE: The location where this marble will store its UUID

Set OE_SIMULATION=1 to run in simulation mode.`

	case "env":
		s = `env ...

Run a command within the ego environment. For example, run
` + me + ` env make
to build a Go project that uses a Makefile.`
	case "signerid":
		s = `signerid <executable|keyfile>

Print the signerID either from the executable or by reading the keyfile.

The keyfile needs to have the extension ".pem"`

	case "uniqueid":
		s = `signerid <executable>

Print the uniqueID from the executable.`

	default:
		s = `<command> [arguments]

Commands:
  sign       Sign an executable built with ego-go.
  run        Run a signed executable.
  env        Run a command in the ego environment.
  marblerun    Run a signed Marblerun marble.
  signerid   Print the signerID of an executable.
  uniqueid   Print the uniqueID of an executable.

Use "` + me + ` help <command>" for more information about a command.`
	}

	fmt.Println("Usage: " + me + " " + s)
}
