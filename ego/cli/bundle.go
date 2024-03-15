// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"archive/tar"
	"compress/gzip"
	"debug/elf"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// ErrFileIsNotAnEGoBinary is returned when the file is not an enclave, therefore cannot possibly be bundled.
var ErrFileIsNotAnEGoBinary = errors.New("file is not an EGo Binary")

// ErrFileIsAlreadyBundled is returned when the file is already bundled. Technically, the outer binary is not an enclave either, but this is a more specific error.
var ErrFileIsAlreadyBundled = errors.New("file is already bundled")

// ErrFileHasNotBeenSignedYet is returned when the enclave has not been signed yet with 'ego sign', which would create a non-functional bundle.
var ErrFileHasNotBeenSignedYet = errors.New("enclave has not been signed yet with 'ego sign'")

// Bundle bundles an enclave with the currently installed EGo runtime into a single executable.
func (c *Cli) Bundle(filename string, outputFilename string) (reterr error) {
	// Check if the file is an enclave or already bundled
	if err := c.checkIfBundable(filename); err != nil {
		return err
	}

	// Build .tar.gz image containing the runtime and user enclave
	tarFilename, err := c.buildImage(filename)
	if err != nil {
		return err
	}
	defer func() { _ = c.fs.Remove(tarFilename) }()

	if outputFilename == "" {
		outputFilename = filepath.Base(filename) + "-bundle"
	}

	// Prepare the bundle
	if err := c.prepareBundle(filename, outputFilename); err != nil {
		return err
	}
	defer func() {
		if reterr != nil {
			_ = c.fs.Remove(outputFilename)
		}
	}()

	// Add the runtime image to the bundle
	if err := addSectionToELF(outputFilename, tarFilename, ".ego.bundle"); err != nil {
		return err
	}

	fmt.Printf("Saved bundle as: %s\n", outputFilename)

	return nil
}

// checkIfBundable checks if the file is an enclave or already bundled.
func (c *Cli) checkIfBundable(filename string) error {
	file, err := c.fs.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Open the ELF file and check if a section called ".oeinfo" exists
	elf, err := elf.NewFile(file)
	if err != nil {
		return err
	}

	// Check if the ELF file has a section called ".oeinfo"
	var isEnclave bool
	var isAlreadyBundled bool
	var isGoBinary bool
	for _, section := range elf.Sections {
		switch section.Name {
		case ".oeinfo":
			isEnclave = true
		case ".ego.bundle":
			isAlreadyBundled = true
		case ".go.buildinfo":
			isGoBinary = true
		}
	}

	if isAlreadyBundled {
		return ErrFileIsAlreadyBundled
	}

	if !isEnclave || !isGoBinary {
		return ErrFileIsNotAnEGoBinary
	}

	// Check if the binary has been signed with `ego sign` yet by checking if the payload has been appended.
	// We do not check here if it is actually valid. If it is not valid, the binary can still be bundled, but during launch the enclave will refuse to run.
	payloadSize, _, _, err := getPayloadInformation(file)
	if err != nil {
		return err
	}

	if payloadSize == 0 {
		return ErrFileHasNotBeenSignedYet
	}

	return nil
}

// buildImage builds a gzip-compressed tarball containing the EGo runtime.
func (c *Cli) buildImage(enclaveFilename string) (tempFileName string, reterr error) {
	// Create the tar file
	tempFile, err := c.fs.TempFile("", "ego-runtime")
	if err != nil {
		return "", err
	}
	defer func() {
		if reterr != nil {
			_ = c.fs.Remove(tempFile.Name())
		}
	}()
	defer tempFile.Close()

	// Setup tar writer with gzip compression
	compressedWriter := gzip.NewWriter(tempFile)
	defer compressedWriter.Close()
	tarWriter := tar.NewWriter(compressedWriter)
	defer tarWriter.Close()

	// Add the runtime image to the tar file
	if err := c.addToArchive(tarWriter, c.getEgoHostPath(), "ego-host"); err != nil {
		return "", err
	}
	if err := c.addToArchive(tarWriter, c.getEgoEnclavePath(), "ego-enclave"); err != nil {
		return "", err
	}

	// Add the enclave to the tar file
	if err := c.addToArchive(tarWriter, enclaveFilename, "enclave"); err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

// addToArchive adds a file from the filesystem to the tar archive.
func (c *Cli) addToArchive(tw *tar.Writer, sourceFilename string, targetFilename string) error {
	file, err := c.fs.Open(sourceFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	sourceFileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(sourceFileInfo, sourceFileInfo.Name())
	if err != nil {
		return err
	}

	header.Name = targetFilename

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	n, err := io.Copy(tw, file)
	if err != nil {
		return err
	}

	if n != sourceFileInfo.Size() {
		return fmt.Errorf("file size mismatch: %d written instead of expected %d bytes", n, sourceFileInfo.Size())
	}

	return nil
}

// prepareBundle copies the "empty" bundle executable to the target location and names it after the enclave to be bundled.
func (c *Cli) prepareBundle(inputFilename string, outputFilename string) (reterr error) {
	// Copy the base bundle loader binary into the new file
	loaderBinary, err := c.fs.Open(c.getEgoBundlePath())
	if err != nil {
		return err
	}
	defer loaderBinary.Close()

	loaderBinaryInfo, err := loaderBinary.Stat()
	if err != nil {
		return err
	}

	// Create target bundled file. Apparently, 'go build' creates binaries with 775... So let's just replicate this.
	binaryToPack, err := c.fs.OpenFile(outputFilename, os.O_CREATE|os.O_RDWR, 0o775)
	if err != nil {
		return err
	}
	defer func() {
		if reterr != nil {
			_ = c.fs.Remove(binaryToPack.Name())
		}
	}()
	defer binaryToPack.Close()

	n, err := io.Copy(binaryToPack, loaderBinary)
	if err != nil {
		return err
	}
	if n != loaderBinaryInfo.Size() {
		return fmt.Errorf("file size mismatch: %d written instead of expected %d bytes", n, loaderBinaryInfo.Size())
	}

	return nil
}

func (c *Cli) getEgoBundlePath() string {
	return filepath.Join(c.egoPath, "share", "ego-bundle")
}

// addSectionToELF calls objcopy (part of binutils) to add a local binary file as a section to an ELF file.
func addSectionToELF(targetFilename string, dataFilename string, sectionName string) error {
	// Use objcopy to copy the data file into the target ELF file
	cmd := exec.Command("objcopy", "--add-section", sectionName+"="+dataFilename, targetFilename, targetFilename)
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Println(string(exitError.Stderr))
			return fmt.Errorf("objcopy exited with error: %d", exitError.ExitCode())
		}
		return err
	}

	return nil
}
