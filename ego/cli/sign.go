// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"debug/elf"
	"ego/config"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const defaultConfigFilename = "enclave.json"
const defaultPrivKeyFilename = "private.pem"
const defaultPubKeyFilename = "public.pem"

// ErrNoOEInfo defines an error when no .oeinfo section could be found. This likely occures whend the binary to sign was not built with ego-go.
var ErrNoOEInfo = errors.New("could not find .oeinfo section")

func (c *Cli) signWithJSON(conf *config.Config) error {
	//write temp .conf file
	cProduct := "ProductID=" + strconv.Itoa(conf.ProductID) + "\n"
	cSecurityVersion := "SecurityVersion=" + strconv.Itoa(conf.SecurityVersion) + "\n"

	var cDebug string
	if conf.Debug {
		cDebug = "Debug=1\n"
	} else {
		cDebug = "Debug=0\n"
	}

	// calculate number of pages: HeapSize[MiB], pageSize is 4096B
	heapPages := conf.HeapSize * 1024 * 1024 / 4096
	cNumHeapPages := "NumHeapPages=" + strconv.Itoa(heapPages) + "\n"

	cStackPages := "NumStackPages=1024\n"
	cNumTCS := "NumTCS=32\n"

	var cExecutableHeap string
	if conf.ExecutableHeap {
		cExecutableHeap = "ExecutableHeap=1\n"
	}

	file, err := c.fs.TempFile("", "")
	if err != nil {
		return err
	}
	defer c.fs.Remove(file.Name())

	_, err = file.Write([]byte(cProduct + cSecurityVersion + cDebug + cNumHeapPages + cStackPages + cNumTCS + cExecutableHeap))
	if err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	// Prepare JSON data for embedding to the executable
	jsonData, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	// Embed enclave.json inside executable as payload
	if err := c.embedConfigAsPayload(conf.Exe, jsonData); err != nil {
		return err
	}

	//create public and private key if private key does not exits
	c.createDefaultKeypair(conf.Key)

	enclavePath := filepath.Join(c.egoPath, "share", "ego-enclave")
	cmd := exec.Command(c.getOesignPath(), "sign", "-e", enclavePath, "-c", file.Name(), "-k", conf.Key, "--payload", conf.Exe)
	out, err := c.runner.CombinedOutput(cmd)
	if _, ok := err.(*exec.ExitError); ok {
		return errors.New(string(out))
	}
	return err
}

func (c *Cli) signExecutable(path string) error {
	conf, err := c.readConfigJSONtoStruct(defaultConfigFilename)

	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else if conf.Exe == path {
		return c.signWithJSON(conf)
	} else {
		return fmt.Errorf("provided path to executable does not match the one in enclave.json")
	}

	//sane default values
	conf = &config.Config{
		Exe:             path,
		Key:             defaultPrivKeyFilename,
		Debug:           true,
		HeapSize:        512, //[MB]
		ProductID:       1,
		SecurityVersion: 1,
	}

	jsonData, err := json.MarshalIndent(conf, "", " ")
	if err != nil {
		return err
	}
	if err := c.fs.WriteFile(defaultConfigFilename, jsonData, 0644); err != nil {
		return err
	}

	return c.signWithJSON(conf)
}

// Reads the provided File and turns it into a struct
// after some basic sanity check are performed it is returned
// err != nil indicates that the file could not be read or the
// JSON could not be unmarshalled
func (c *Cli) readConfigJSONtoStruct(path string) (*config.Config, error) {
	data, err := c.fs.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var conf config.Config
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	if err := conf.PopulateContent(c.fs); err != nil {
		return nil, err
	}
	return &conf, nil
}

// Creates a public/secret keypair if the provided secret key does not exists
func (c *Cli) createDefaultKeypair(file string) {
	if _, err := c.fs.Stat(file); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		fmt.Println("Generating new " + file)
		// SGX requires the RSA exponent to be 3. Go's API does not support this.
		if err := c.runner.Run(exec.Command("openssl", "genrsa", "-out", file, "-3", "3072")); err != nil {
			panic(err)
		}
		pubPath := filepath.Join(filepath.Dir(file), defaultPubKeyFilename)
		if err := c.runner.Run(exec.Command("openssl", "rsa", "-in", file, "-pubout", "-out", pubPath)); err != nil {
			panic(err)
		}
	}
}

// Sign signs an executable built with ego-go.
func (c *Cli) Sign(filename string) error {
	if filename == "" {
		conf, err := c.readConfigJSONtoStruct(defaultConfigFilename)
		if err != nil {
			return err
		}
		return c.signWithJSON(conf)
	}
	if filepath.Ext(filename) == ".json" {
		conf, err := c.readConfigJSONtoStruct(filename)
		if err != nil {
			return err
		}
		return c.signWithJSON(conf)
	}
	return c.signExecutable(filename)
}

func (c *Cli) embedConfigAsPayload(path string, jsonData []byte) error {
	// Load ELF executable
	f, err := c.fs.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	// Check if a payload already exists
	payloadSize, payloadOffset, oeInfoOffset, err := getPayloadInformation(f)
	if err != nil {
		return err
	}

	// If a payload already exists, truncate the file to remove it
	if payloadSize > 0 {
		fileStat, err := f.Stat()
		if err != nil {
			return err
		}

		// Check if payload is at expected location
		expectedPayloadOffset := fileStat.Size() - int64(payloadSize)
		if expectedPayloadOffset != payloadOffset {
			return errors.New("expected payload location does not match real payload location, cannot safely truncate old payload")
		}

		err = f.Truncate(payloadOffset)
		if err != nil {
			return err
		}
	} else if (payloadSize == 0) != (payloadOffset == 0) {
		return errors.New("payload information in header is inconsistent, cannot continue")
	}

	// Get current file size to determine offset
	fileStat, err := f.Stat()
	if err != nil {
		return err
	}
	filesize := fileStat.Size()

	// Write payload offset and size to .oeinfo header
	if err := writeUint64At(f, uint64(filesize), oeInfoOffset+2048); err != nil {
		return err
	}
	if err := writeUint64At(f, uint64(len(jsonData)), oeInfoOffset+2056); err != nil {
		return err
	}

	// And finally, append the payload to the file
	n, err := f.WriteAt(jsonData, filesize)
	if err != nil {
		return err
	} else if n != len(jsonData) {
		return errors.New("failed to embed enclave.json as metadata")
	}

	return nil
}

func getPayloadInformation(f io.ReaderAt) (uint64, int64, int64, error) {
	// .oeinfo + 2056 contains the size of an embedded Edgeless RT data payload.
	// If it is > 0, a payload already exists.

	elfFile, err := elf.NewFile(f)
	if err != nil {
		return 0, 0, 0, err
	}

	oeInfo := elfFile.Section(".oeinfo")
	if oeInfo == nil {
		return 0, 0, 0, ErrNoOEInfo
	}

	payloadOffset, err := readUint64At(oeInfo, 2048)
	if err != nil {
		return 0, 0, 0, err
	}
	payloadSize, err := readUint64At(oeInfo, 2056)
	if err != nil {
		return 0, 0, 0, err
	}

	return payloadSize, int64(payloadOffset), int64(oeInfo.Offset), nil
}

func writeUint64At(w io.WriterAt, x uint64, off int64) error {
	xByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(xByte, x)

	n, err := w.WriteAt(xByte, off)
	if err != nil {
		return err
	} else if n != 8 {
		return errors.New("did not write expected number of bytes")
	}

	return nil
}

func readUint64At(r io.ReaderAt, off int64) (uint64, error) {
	xByte := make([]byte, 8)

	n, err := r.ReadAt(xByte, off)
	if err != nil {
		return 0, err
	} else if n != 8 {
		return 0, errors.New("did not read expected number of bytes")
	}

	return binary.LittleEndian.Uint64(xByte), nil
}
