// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"debug/elf"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// create an unsigned EGo executable
var elfUnsigned = func() []byte {
	const outFile = "hello"
	const srcFile = outFile + ".go"

	goroot, err := filepath.Abs(filepath.Join("..", "..", "..", "_ertgo"))
	if err != nil {
		panic(err)
	}

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	// write minimal source file
	const src = `package main;import _"time";func main(){}`
	if err := os.WriteFile(filepath.Join(dir, srcFile), []byte(src), 0o400); err != nil {
		panic(err)
	}

	// compile
	cmd := exec.Command(filepath.Join(goroot, "bin", "go"), "build", srcFile)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOROOT="+goroot)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	// read resulting executable
	data, err := os.ReadFile(filepath.Join(dir, outFile))
	if err != nil {
		panic(err)
	}

	return data
}()

func setupTest(t *testing.T) (string, error) {
	// Setup a fake filesystem to create a tar file
	fs := afero.NewMemMapFs()
	memfsTarFile, err := fs.Create("test-tar-file")
	if err != nil {
		return "", err
	}

	// Create a fake tar file
	gw := gzip.NewWriter(memfsTarFile)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	if err := addStubFileToArchive(tw, fs, "bin/ego-host"); err != nil {
		return "", err
	}
	if err := addStubFileToArchive(tw, fs, "share/ego-enclave"); err != nil {
		return "", err
	}
	if err := addStubFileToArchive(tw, fs, "enclave"); err != nil {
		return "", err
	}
	if err := tw.Close(); err != nil {
		return "", err
	}
	if err := gw.Close(); err != nil {
		return "", err
	}

	// Get file size
	memFsTarFileStat, err := memfsTarFile.Stat()
	if err != nil {
		return "", err
	}

	// Reset file offset to so we can copy the file
	_, err = memfsTarFile.Seek(0, 0)
	if err != nil {
		return "", err
	}

	// Copy the tar file to the host
	osTarFile, err := os.CreateTemp("", "ego-unittest")
	if err != nil {
		return "", err
	}
	defer osTarFile.Close()
	defer os.Remove(osTarFile.Name())

	n, err := io.Copy(osTarFile, memfsTarFile)
	if err != nil {
		return osTarFile.Name(), err
	} else if n != memFsTarFileStat.Size() {
		return osTarFile.Name(), fmt.Errorf("copied %d bytes, expected %d", n, memFsTarFileStat.Size())
	}

	// Copy our silently compiled test-binary, which is not the real bundle executable, to the host file system
	osBinaryFile, err := os.CreateTemp("", "ego-unittest")
	if err != nil {
		return osTarFile.Name(), err
	}
	defer osBinaryFile.Close()

	n, err = io.Copy(osBinaryFile, bytes.NewReader(elfUnsigned))
	if err != nil {
		return osBinaryFile.Name(), err
	} else if n != int64(len(elfUnsigned)) {
		return osBinaryFile.Name(), fmt.Errorf("copied %d bytes, expected %d", n, len(elfUnsigned))
	}

	// Embed the tar file into the elf file
	if err := addSectionToELF(osBinaryFile.Name(), osTarFile.Name(), ".ego.bundle"); err != nil {
		return osBinaryFile.Name(), err
	}

	return osBinaryFile.Name(), nil
}

func TestUnpackBundledImage(t *testing.T) {
	objdumpPath, err := exec.LookPath("objdump")
	if err != nil || objdumpPath == "" {
		t.Skip("objdump not found, cannot run this test.")
	}

	assert := assert.New(t)
	require := require.New(t)

	// Create test files
	elfFilePath, err := setupTest(t)
	if elfFilePath != "" {
		defer os.Remove(elfFilePath)
	}
	require.NoError(err)

	// Load the ELF file
	elfFile, err := elf.Open(elfFilePath)
	require.NoError(err)
	defer elfFile.Close()

	// Run the unpacker
	fs := afero.NewMemMapFs()
	unpackingPath, err := afero.TempDir(fs, "", "ego-unittest")
	require.NoError(err)
	defer func() { _ = fs.RemoveAll(unpackingPath) }()

	assert.NoError(unpackBundledImage(fs, unpackingPath, elfFile))

	// Check that the tar file was unpacked
	egoHostFile, err := afero.ReadFile(fs, filepath.Join(unpackingPath, "bin/ego-host"))
	assert.NoError(err)
	assert.Equal(string(egoHostFile), "this is a test")
	egoEnclaveFile, err := afero.ReadFile(fs, filepath.Join(unpackingPath, "share/ego-enclave"))
	assert.NoError(err)
	assert.Equal(string(egoEnclaveFile), "this is a test")
	enclaveFile, err := afero.ReadFile(fs, filepath.Join(unpackingPath, "enclave"))
	assert.NoError(err)
	assert.Equal(string(enclaveFile), "this is a test")
}

func TestRun(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Create test files
	elfFilePath, err := setupTest(t)
	if elfFilePath != "" {
		defer os.Remove(elfFilePath)
	}
	require.NoError(err)

	// Load the ELF file
	elfFile, err := elf.Open(elfFilePath)
	require.NoError(err)
	defer elfFile.Close()

	// Prepare test environment
	fs := afero.NewMemMapFs()
	runner := &testRunner{}
	os.Args = []string{"fake-argv0"}

	// Test to run the main part of the code.
	exitCode, err := run(fs, elfFile, runner)
	assert.NoError(err)
	assert.Equal(42, exitCode) // That's just here to check that this should be the error code returned by the test runner and not by another error inside error.

	/*
		A valid run looks something like this: [/usr/bin/stdbuf -oL /tmp/ego-bundle905847970/bin/ego-host /tmp/ego-bundle905847970/share/ego-enclave:/tmp/ego-bundle905847970/enclave]
		Given that the temp directory is created and deleted inside run, the files are already gone when we exit.
		Therefore, we just check if the paths looks reasonable.
		Correct unpacking should be tested by TestUnpackBundledImage otherwise.
		Also, this uses a weird syntax given that runner.run is a []*exec.Cmd.
		This means we want to take a look at the first run [0], and then check if the thrird [2] and fourth argument [3] contain something like the valid output mentioned above.
	*/
	assert.Contains(runner.run[0].Args[2], "/ego-host")
	assert.Contains(runner.run[0].Args[3], "/ego-enclave:")
	assert.Contains(runner.run[0].Args[3], "/enclave")
}

func addStubFileToArchive(tw *tar.Writer, fs afero.Fs, targetFilename string) error {
	path := filepath.Dir(targetFilename)
	if err := fs.MkdirAll(path, 0o755); err != nil {
		return err
	}

	if err := afero.WriteFile(fs, targetFilename, []byte("this is a test"), 0o755); err != nil {
		return err
	}

	file, err := fs.Open(targetFilename)
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

// Lifted from cli/runner_test.go
type testRunner struct {
	run []*exec.Cmd
}

func (r *testRunner) Run(cmd *exec.Cmd) error {
	r.run = append(r.run, cmd)
	return nil
}

func (*testRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	panic(cmd.Path)
}

func (*testRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	panic(cmd.Path)
}

func (*testRunner) ExitCode(cmd *exec.Cmd) int {
	return 42
}
