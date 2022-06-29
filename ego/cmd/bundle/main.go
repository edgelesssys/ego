// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build linux
// +build linux

package main

import (
	"debug/elf"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"ego/internal/launch"

	"github.com/spf13/afero"
)

// Don't touch! Automatically injected at build-time.
var (
	version   = "0.0.0"
	gitCommit = "0000000000000000000000000000000000000000"
)

// ErrExecutableDoesNotContainBundle is triggered when this executable is launched without containing a bundled enclave image as the ".ego.bundle" section.
var ErrExecutableDoesNotContainBundle = errors.New("executable does not contain an enclave bundle")

func main() {
	fmt.Fprintf(os.Stderr, "EGo (bundled) v%v (%v)\n", version, gitCommit)

	// Open the current execeutable
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	// Parse ourselves as an ELF file so we get access to ELF sections later
	elfFile, err := elf.Open(execPath)
	if err != nil {
		panic(err)
	}
	defer elfFile.Close()

	// Use the "host" environment and not a mocked one for unit tests
	fs := afero.NewOsFs()
	runner := launch.OsRunner{}

	// Run the unpacking & launch the unpacked enclave.
	// Take the error from the enclave as the exit code, unless we failed to execute the enclave before.
	exitCode, err := run(fs, elfFile, runner)
	if err != nil {
		panic(err)
	}
	os.Exit(exitCode)
}

func run(fs afero.Fs, selfElfFile *elf.File, runner launch.Runner) (int, error) {
	// Create a temporary directory for both runtime and enclave
	tempEGoRootPath, err := afero.TempDir(fs, "", "ego-bundle")
	if err != nil {
		return 1, err
	}
	defer fs.RemoveAll(tempEGoRootPath)

	// Register cleanup handler to clean-up on STRG+C
	cleanupHandler(tempEGoRootPath)

	// Unpack the runtime and enclave
	if err := unpackBundledImage(fs, tempEGoRootPath, selfElfFile); err != nil {
		return 1, err
	}

	enclaveFilename, egoHostFilename, egoEnclaveFilename := createPathsForLaunch(tempEGoRootPath)

	// Run the enclave
	exitCode, err := launch.RunEnclave(enclaveFilename, os.Args[1:], egoHostFilename, egoEnclaveFilename, runner)
	if err != nil {
		return 1, err
	}

	return exitCode, nil
}

func createPathsForLaunch(tempEGoRootPath string) (enclaveFilename string, egoHostFilename string, egoEnclaveFilename string) {
	enclaveFilename = filepath.Join(tempEGoRootPath, "enclave")
	egoHostFilename = filepath.Join(tempEGoRootPath, "ego-host")
	egoEnclaveFilename = filepath.Join(tempEGoRootPath, "ego-enclave")
	return
}

// unpackBundledImage decompresses the runtime and unpacks it to the temporary directory.
func unpackBundledImage(fs afero.Fs, path string, elfFile *elf.File) error {
	// Read the runtime section
	runtimeSection := elfFile.Section(".ego.bundle")
	if runtimeSection == nil {
		return ErrExecutableDoesNotContainBundle
	}

	return launch.UntarGzip(fs, runtimeSection.Open(), path)
}

// cleanupHandler removes the temp directory when the exit of the process is requested, since the defered functions will not run otherwise.
func cleanupHandler(egoTempPath string) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.RemoveAll(egoTempPath)
	}()
}
