// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"bufio"
	"ego/cli"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/klauspost/cpuid/v2"
	"github.com/spf13/afero"
)

// Don't touch! Automatically injected at build-time.
var version = "0.0.0"
var gitCommit = "0000000000000000000000000000000000000000"

func main() {
	fmt.Fprintf(os.Stderr, "EGo v%v (%v)\n", version, gitCommit)
	if len(os.Args) < 2 {
		help("")
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]
	c := cli.NewCli(runner{}, afero.NewOsFs())

	switch cmd {
	case "sign":
		var filename string
		if len(args) > 0 {
			filename = args[0]
		}
		err := c.Sign(filename)
		if err == nil {
			return
		}
		if err == cli.ErrNoOEInfo {
			log.Fatalln("ERROR: The .oeinfo section is missing in the binary.\nMaybe the binary was not built with 'ego-go build'?")
		}
		fmt.Println(err)
		if !os.IsNotExist(err) {
			// print error only
			os.Exit(1)
		}
		// also print usage
	case "run":
		if len(args) > 0 {
			exitCode, err := c.Run(args[0], args[1:])
			handleErr(err)
			os.Exit(exitCode)
		}
	case "marblerun":
		if len(args) == 1 {
			exitCode, err := c.Marblerun(args[0])
			handleErr(err)
			os.Exit(exitCode)
		}
	case "signerid":
		if len(args) == 1 {
			id, err := c.Signerid(args[0])
			if err == cli.ErrOECrypto {
				log.Fatalf("ERROR: signerid failed with %v.\nMake sure to pass a valid public key.\n", err)
			}
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(id)
			return
		}
	case "uniqueid":
		if len(args) == 1 {
			id, err := c.Uniqueid(args[0])
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(id)
			return
		}
	case "env":
		if len(args) > 0 {
			os.Exit(c.Env(args[0], args[1:]))
		}
	case "install":
		var component string
		if len(args) == 1 {
			component = args[0]
		}
		if err := c.Install(askInstall, component); err != nil {
			log.Fatal(err)
		}
		return
	case "help":
		if len(args) == 1 {
			help(args[0])
			return
		}
	}

	help(cmd)
}

func handleErr(err error) {
	switch err {
	case nil:
	case cli.ErrElfNoPie:
		fmt.Println("ERROR: failed to load the binary")
		fmt.Println("The binary doesn't seem to be built with 'ego-go build'")
	case cli.ErrValidAttr0:
		fmt.Println("ERROR: failed to load the binary")
		fmt.Println("Please sign the binary with 'ego sign'")
	case cli.ErrEnclIniFailInvalidMeasurement:
		fmt.Println("ERROR: failed to initialize the enclave")
		fmt.Println("Try to resign the binary with 'ego sign' and rerun afterwards.")
	case cli.ErrEnclIniFailUnexpected:
		fmt.Println("ERROR: failed to initialize the enclave")
		if _, err := os.Stat("/dev/isgx"); err == nil {
			fmt.Println("Try to run: sudo ego install libsgx-launch")
		}
	case cli.ErrSGXOpenFail:
		fmt.Println("ERROR: failed to open Intel SGX device")
		if cpuid.CPU.Supports(cpuid.SGX) {
			fmt.Println("Install the SGX driver with: sudo ego install sgx-driver")
		} else {
			fmt.Println("This machine doesn't support SGX.")
		}
		fmt.Println("You can use 'OE_SIMULATION=1 ego run ...' to run in simulation mode.")
	default:
		fmt.Println(err)
	}
}

func help(cmd string) {
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
ego env make
to build a Go project that uses a Makefile.`

	case "signerid":
		s = `signerid <executable | key.pem>

Print the SignerID either from a signed executable or by reading a keyfile.`

	case "uniqueid":
		s = `uniqueid <executable>

Print the UniqueID of a signed executable.`

	case "install":
		s = `install [component]

Install drivers and other components. The components that you can install depend on your operating system and its version.
Use "ego install" to list the available components for your system.`

	default:
		s = `<command> [arguments]

Commands:
  sign        Sign an executable built with ego-go.
  run         Run a signed executable in standalone mode.
  marblerun   Run a signed executable as a Marblerun Marble.
  signerid    Print the SignerID of a signed executable.
  uniqueid    Print the UniqueID of a signed executable.
  env         Run a command in the EGo environment.
  install     Install drivers and other components.

Use "ego help <command>" for more information about a command.`
	}

	fmt.Println("\nUsage: ego " + s)
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
