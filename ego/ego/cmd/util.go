// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"fmt"
	"os"

	"github.com/edgelesssys/ego/ego/cli"
	"github.com/edgelesssys/ego/ego/internal/launch"
	"github.com/klauspost/cpuid/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newCli() *cli.Cli {
	return cli.NewCli(launch.OsRunner{}, afero.NewOsFs())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func hideHelpFlag(cmd *cobra.Command) {
	// must init help flag before it can be hidden
	cmd.InitDefaultHelpFlag()
	must(cmd.Flags().MarkHidden("help"))
}

func handleErr(err error) {
	switch err {
	case nil:
	case launch.ErrElfNoPie:
		fmt.Println("ERROR: failed to load the binary")
		fmt.Println("The binary doesn't seem to be built with 'ego-go build'")
	case launch.ErrValidAttr0:
		fmt.Println("ERROR: failed to load the binary")
		fmt.Println("Please sign the binary with 'ego sign'")
	case launch.ErrEnclIniFailInvalidMeasurement:
		fmt.Println("ERROR: failed to initialize the enclave")
		fmt.Println("Try to resign the binary with 'ego sign' and rerun afterwards.")
	case launch.ErrEnclIniFailUnexpected:
		fmt.Println("ERROR: failed to initialize the enclave")
		if _, err := os.Stat("/dev/isgx"); err == nil {
			fmt.Println("Try to run: sudo ego install libsgx-launch")
		}
	case launch.ErrSGXOpenFail:
		fmt.Println("ERROR: failed to open Intel SGX device")
		if cpuid.CPU.Supports(cpuid.SGX) {
			// the error is also thrown if /dev/sgx_enclave exists, but /dev/sgx does not
			if _, err := os.Stat("/dev/sgx_enclave"); err == nil {
				// libsgx-enclave-common will create the symlinks
				fmt.Println("Install the SGX base package with: sudo ego install libsgx-enclave-common")
			} else {
				fmt.Println("Install the SGX driver with: sudo ego install sgx-driver")
			}
		} else {
			fmt.Println("This machine doesn't support SGX.")
		}
		fmt.Println("You can use 'OE_SIMULATION=1 ego run ...' to run in simulation mode.")
	case launch.ErrLoadDataFailUnexpected:
		fmt.Println("ERROR: failed to initialize the enclave")
		fmt.Println("Install the SGX base package with: sudo ego install libsgx-enclave-common")
		fmt.Println("Or temporarily fix the error with: sudo mount -o remount,exec /dev")
	case launch.ErrOECrypto:
		fmt.Printf("ERROR: signerid failed with %v.\nMake sure to pass a valid public key.\n", err)
	case cli.ErrNoOEInfo:
		fmt.Println("ERROR: The .oeinfo section is missing in the binary.\nMaybe the binary was not built with 'ego-go build'?")
	case cli.ErrUnsupportedImportEClient:
		fmt.Println("ERROR: You cannot import the github.com/edgelesssys/ego/eclient package within the EGo enclave.")
		fmt.Println("It is intended to be used for applications running outside the SGX enclave.")
		fmt.Println("You can use the github.com/edgelesssys/ego/enclave package as a replacement for usage inside the enclave.")
	case cli.ErrLargeHeapWithSmallHeapSize:
		fmt.Println("ERROR: The binary is built with build tag \"ego_largeheap\", but heapSize is set to less than 512.")
		fmt.Println("Either increase heapSize or rebuild without this build tag.")
	case cli.ErrNoLargeHeapWithLargeHeapSize:
		fmt.Println("ERROR: The binary is built without build tag \"ego_largeheap\", but heapSize is set to more than 16384.")
		fmt.Println("Either decrease heapSize or rebuild with this build tag:")
		fmt.Println("\tego-go build -tags ego_largeheap ...")
	default:
		fmt.Println("ERROR:", err)
	}
}
