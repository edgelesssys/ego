// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"debug/elf"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/ego/ego/config"
	"github.com/edgelesssys/ego/ego/internal/launch"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckIfBundable(t *testing.T) {
	/*
		Does not test the following cases:
		- OpenEnclave binary, but not an EGo enclave
		- already-bundled EGo binary
	*/
	const exe = "to-be-signed-enclave"

	assert := assert.New(t)
	require := require.New(t)

	// Setup a fake filesystem and fake signer.
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	runner := signRunner{fs: fs}
	cli := NewCli(&runner, fs)

	// First, check a no enclave binary. Let's use ourselves as a test here, so we do not have any special dependencies.
	testNonEnclaveBinaryPath, err := os.Executable()
	require.NoError(err)
	testNonEnclaveBinary, err := os.ReadFile(testNonEnclaveBinaryPath)
	require.NoError(err)

	require.NoError(fs.WriteFile("no-enclave", testNonEnclaveBinary, 0o644))
	assert.ErrorIs(cli.checkIfBundable("no-enclave"), ErrFileIsNotAnEGoBinary)

	// Now, let's check an unsigned binary
	require.NoError(fs.WriteFile(exe, elfUnsigned, 0o644))
	assert.ErrorIs(cli.checkIfBundable(exe), ErrFileHasNotBeenSignedYet)

	// Let's fake-sign the binary by just embedding the enclave configuration.
	testConf := &config.Config{
		Exe:             exe,
		Key:             defaultPrivKeyFilename,
		Debug:           true,
		HeapSize:        512, //[MB]
		ProductID:       1,
		SecurityVersion: 1,
	}
	jsonData, err := json.Marshal(testConf)
	require.NoError(err)
	err = cli.embedConfigAsPayload(exe, jsonData)
	assert.NoError(err)

	// Now, check if our enclave passes as bundable
	assert.NoError(cli.checkIfBundable(exe))
}

func TestAddToArchive(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Setup a fake filesystem
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	cli := NewCli(nil, fs)
	require.NoError(fs.MkdirAll("/this/is/a/directory/test", 0o777))
	require.NoError(fs.WriteFile("/this/is/a/directory/test/test.txt", []byte("test"), 0o644))
	require.NoError(fs.WriteFile("/rootfile.txt", []byte("test"), 0o644))
	tmpTarFile, err := fs.TempFile("", "")
	require.NoError(err)

	// Setup tar writer with gzip compression
	compressedWriter := gzip.NewWriter(tmpTarFile)
	defer compressedWriter.Close()
	tarWriter := tar.NewWriter(compressedWriter)
	defer tarWriter.Close()

	// Add files to archive
	assert.NoError(cli.addToArchive(tarWriter, "/this/is/a/directory/test/test.txt", "another/directory/test.txt"))
	assert.NoError(cli.addToArchive(tarWriter, "/rootfile.txt", "unpacked-rootfile.txt"))
	tarWriter.Close()
	compressedWriter.Close()

	// Try to untar the archive
	tarFile, err := fs.Open(tmpTarFile.Name())
	require.NoError(err)
	assert.NoError(launch.UntarGzip(fs, tarFile, "/unpacked"))

	// Check if the files are there
	unpackedRootFileContent, err := fs.ReadFile("/unpacked/unpacked-rootfile.txt")
	assert.NoError(err)
	assert.Equal([]byte("test"), unpackedRootFileContent)
	unpackedDirectoryFileContent, err := fs.ReadFile("/unpacked/another/directory/test.txt")
	assert.NoError(err)
	assert.Equal([]byte("test"), unpackedDirectoryFileContent)
}

func TestBuildImage(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Setup a fake filesystem
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	cli := NewCli(nil, fs)

	fakeContentEgoHost := []byte("this-is-no-real-elf: ego-host")
	fakeContentEgoEnclave := []byte("this-is-no-real-elf: ego-enclave")
	fakeContentMyEnclave := []byte("this-is-no-real-elf: my-enclave")

	binPath := filepath.Join(cli.egoPath, "bin")
	sharePath := filepath.Join(cli.egoPath, "share")
	require.NoError(fs.MkdirAll(binPath, 0o777))
	require.NoError(fs.MkdirAll(sharePath, 0o777))
	require.NoError(fs.WriteFile(filepath.Join(binPath, "ego-host"), fakeContentEgoHost, 0o644))
	require.NoError(fs.WriteFile(filepath.Join(sharePath, "ego-enclave"), fakeContentEgoEnclave, 0o644))
	require.NoError(fs.WriteFile("my-enclave", fakeContentMyEnclave, 0o644))

	// Run the prepare bundle command
	pathToBundle, err := cli.buildImage("my-enclave")
	assert.NoError(err)

	bundleFile, err := fs.Open(pathToBundle)
	require.NoError(err)

	/*
		Untar the archive. The structure should be the following:
		- bin/ego-host
		- share/ego-enclave
		- my-enclave
	*/
	require.NoError(launch.UntarGzip(fs, bundleFile, "/unpacked"))
	unpackedBinFileContent, err := fs.ReadFile("/unpacked/ego-host")
	assert.NoError(err)
	assert.Equal(fakeContentEgoHost, unpackedBinFileContent)
	unpackedShareFileContent, err := fs.ReadFile("/unpacked/ego-enclave")
	assert.NoError(err)
	assert.Equal(fakeContentEgoEnclave, unpackedShareFileContent)
	unpackedEnclaveFileContent, err := fs.ReadFile("/unpacked/enclave")
	assert.NoError(err)
	assert.Equal(fakeContentMyEnclave, unpackedEnclaveFileContent)
}

func TestPrepareBundle(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Setup a fake filesystem
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	cli := NewCli(nil, fs)

	fakeContentEgoBundle := []byte("this-is-no-real-elf: ego-bundle")
	sharePath := filepath.Join(cli.egoPath, "share")
	require.NoError(fs.MkdirAll(sharePath, 0o777))
	require.NoError(fs.WriteFile(filepath.Join(sharePath, "ego-bundle"), fakeContentEgoBundle, 0o644))
	file, err := fs.Create("my-enclave")
	require.NoError(err)
	defer file.Close()

	// Run the prepare bundle command
	assert.NoError(cli.prepareBundle(file.Name(), "created-bundle"))

	// Check if the bundle has been copied to the right place
	bundleContent, err := fs.ReadFile("created-bundle")
	assert.NoError(err)
	assert.Equal(fakeContentEgoBundle, bundleContent)
}

func TestAddSectionToELF(t *testing.T) {
	// This test stores files on the real filesystem, as we cannot get away with mocking here due to the external call to objdump which requires a real file...
	// If objdump does not exist on the current running system, we skip this step.
	objdumpPath, err := exec.LookPath("objdump")
	if err != nil || objdumpPath == "" {
		t.Skip("objdump not found, cannot run this test.")
	}

	assert := assert.New(t)
	require := require.New(t)

	// Create two test files, with the second one to be embedded into the first one
	testElfFile, err := os.CreateTemp("", "ego-unittest")
	require.NoError(err)
	defer testElfFile.Close()
	defer os.Remove(testElfFile.Name())
	testContentFile, err := os.CreateTemp("", "ego-unittest")
	require.NoError(err)
	defer testElfFile.Close()
	defer os.Remove(testElfFile.Name())

	// Write the elf file which is supposed to get the section added
	n, err := io.Copy(testElfFile, bytes.NewReader(elfUnsigned))
	require.NoError(err)
	require.EqualValues(len(elfUnsigned), n)

	// Write the content of the section to be added
	testContent := []byte("this is a test")
	n, err = io.Copy(testContentFile, bytes.NewReader(testContent))
	require.NoError(err)
	require.EqualValues(len(testContent), n)

	// Embed the content of the content file into the elf file
	assert.NoError(addSectionToELF(testElfFile.Name(), testContentFile.Name(), ".ego.bundle"))

	// Check if the section has been added to the elf file
	elf, err := elf.Open(testElfFile.Name())
	require.NoError(err)
	defer elf.Close()
	elfSection := elf.Section(".ego.bundle")
	require.NotNil(elfSection)

	// Check if the section has the correct content
	sectionContent, err := elfSection.Data()
	assert.NoError(err)
	assert.Equal(testContent, sectionContent)
}
