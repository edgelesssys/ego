// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"debug/elf"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/spf13/afero"
)

const defaultConfigFilename = "enclave.json"
const defaultPrivKeyFilename = "private.pem"
const defaultPubKeyFilename = "public.pem"

type config struct {
	Exe             string `json:"exe"`
	Key             string `json:"key"`
	Debug           bool   `json:"debug"`
	HeapSize        int    `json:"heapSize"`
	ProductID       int    `json:"productID"`
	SecurityVersion int    `json:"securityVersion"`
}

// Validate Exe, Key, HeapSize
func (c *config) validate() error {
	if c.HeapSize == 0 {
		return fmt.Errorf("heapSize not set in config file")
	}
	if c.Exe == "" {
		return fmt.Errorf("exe not set in config file")
	}
	if c.Key == "" {
		return fmt.Errorf("key not set in config file")
	}
	return nil
}

func (c *Cli) signWithJSON(conf *config) error {
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

	file, err := c.fs.TempFile("", "")
	if err != nil {
		return err
	}
	defer c.fs.Remove(file.Name())

	_, err = file.Write([]byte(cProduct + cSecurityVersion + cDebug + cNumHeapPages + cStackPages + cNumTCS))
	if err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	// Prepare JSON data for embedding to the executable
	jsonData, err := json.MarshalIndent(conf, "", " ")
	if err != nil {
		return err
	}

	// Embed enclave.json inside executable as payload
	c.embedConfigAsPayload(conf.Exe, jsonData)

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
		return fmt.Errorf("Provided path to executable does not match the one in enclave.json")
	}

	//sane default values
	conf = &config{
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
func (c *Cli) readConfigJSONtoStruct(path string) (*config, error) {
	data, err := c.fs.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var conf config
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	if err := conf.validate(); err != nil {
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
	f, err := c.fs.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Check if a payload already exists
	payloadSize, oeInfoOffset, err := checkIfPayloadExists(f)
	if err != nil {
		return err
	}

	// If a payload already exists, truncate the file to remove it
	if payloadSize > 0 {
		fileStat, err := f.Stat()
		if err != nil {
			return err
		}
		err = f.Truncate(fileStat.Size() - int64(payloadSize))
		if err != nil {
			return err
		}
	}

	// Append the new payload size + offset to the .oeinfo header
	newPayloadSize := make([]byte, 8)
	newPayloadOffset := make([]byte, 8)

	fileStat, err := f.Stat()
	if err != nil {
		return err
	}
	filesize := fileStat.Size()

	binary.LittleEndian.PutUint64(newPayloadSize, uint64(len(jsonData)))
	binary.LittleEndian.PutUint64(newPayloadOffset, uint64(filesize))

	n, err := f.WriteAt(newPayloadOffset, oeInfoOffset+2048)
	if err != nil {
		return err
	} else if n != 8 {
		return errors.New("failed to embed payload metadata to executable")
	}

	n, err = f.WriteAt(newPayloadSize, oeInfoOffset+2056)
	if err != nil {
		return err
	} else if n != 8 {
		return errors.New("failed to embed payload metadata to executable")
	}

	// And finally, append the payload to the file
	n, err = f.WriteAt(jsonData, filesize)
	if err != nil {
		return err
	} else if n != len(jsonData) {
		return errors.New("failed to embed enclave.json as metadata")
	}

	return nil
}

func checkIfPayloadExists(f afero.File) (uint64, int64, error) {
	// .oeinfo + 2056 contains the size of an embedded Edgeless RT data payload.
	// If it is > 0, a payload already exists.

	elfFile, err := elf.NewFile(f)
	if err != nil {
		return 0, 0, err
	}

	oeInfo := elfFile.Section(".oeinfo")
	if oeInfo == nil {
		panic(errors.New("could not find .oeinfo section"))
	}

	payloadSizeBinary := make([]byte, 8)
	n, err := oeInfo.ReadAt(payloadSizeBinary, 2056)
	if err != nil {
		return 0, 0, err
	} else if n != 8 {
		return 0, 0, errors.New("could not read embedded payload size from executable")
	}

	payloadSize := binary.LittleEndian.Uint64(payloadSizeBinary)

	return payloadSize, int64(oeInfo.Offset), nil
}
