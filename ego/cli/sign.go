// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"debug/elf"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/edgelesssys/ego/ego/config"
)

const (
	defaultConfigFilename  = "enclave.json"
	defaultPrivKeyFilename = "private.pem"
	defaultPubKeyFilename  = "public.pem"
)

// ErrNoOEInfo defines an error when no .oeinfo section could be found. This likely occurs when the binary to sign was not built with ego-go.
var ErrNoOEInfo = errors.New("could not find .oeinfo section")

// errConfigDoesNotExist defines an error when enclave.json or the expected .json file from user input does not exist. It is only used for internal error handling.
var errConfigDoesNotExist = errors.New("enclave config file not found")

func (c *Cli) signWithJSON(conf *config.Config) error {
	if symbols, err := c.getSymbolsFromELF(conf.Exe); err == nil {
		// First, check if the executable does not contain unsupported imports / symbols.
		if err := checkUnsupportedImports(symbols); err != nil {
			return err
		}

		// Check that heapSize is in the supported range of the heap mode the binary was built with.
		if err := checkHeapMode(symbols, conf.HeapSize); err != nil {
			return err
		}
	} else if !errors.Is(err, elf.ErrNoSymbols) {
		return fmt.Errorf("getting symbols: %w", err)
	}

	// write temp .conf file
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
	defer func() { _ = c.fs.Remove(file.Name()) }()

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

	// create public and private key if private key does not exist
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
	// Try to parse existing config
	conf, err := c.readConfigJSONtoStruct(defaultConfigFilename)

	// If an enclave.json exists, check if the path inside it matches the user input and call signWithJSON.
	// Otherwise, throw an error.
	if err != nil {
		if !errors.Is(err, errConfigDoesNotExist) {
			return err
		}
	} else if conf.Exe == path {
		return c.signWithJSON(conf)
	} else {
		return fmt.Errorf("provided path to executable does not match the one in %s", defaultConfigFilename)
	}

	// If no enclave.json exists, generate a new one.
	fmt.Println("Generating new", defaultConfigFilename)

	// sane default values
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
	if err := c.fs.WriteFile(defaultConfigFilename, jsonData, 0o644); err != nil {
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
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %v", errConfigDoesNotExist, path)
	}
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

	// interpret paths relative to the directory containing the config file
	dir := filepath.Dir(path)
	conf.Exe = filepath.Join(dir, conf.Exe)
	conf.Key = filepath.Join(dir, conf.Key)

	if err := conf.PopulateContent(c.fs); err != nil {
		return nil, fmt.Errorf("failed to populate embedded file content: %w", err)
	}
	return &conf, nil
}

// Creates a public/secret keypair if the provided secret key does not exist
func (c *Cli) createDefaultKeypair(file string) {
	if _, err := c.fs.Stat(file); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		fmt.Println("Generating new " + file)
		// SGX requires the RSA exponent to be 3. Go's API does not support this.
		if out, err := c.runner.CombinedOutput(exec.Command("openssl", "genrsa", "-out", file, "-3", "3072")); err != nil {
			fmt.Println(string(out))
			panic(err)
		}
		pubPath := filepath.Join(filepath.Dir(file), defaultPubKeyFilename)
		if out, err := c.runner.CombinedOutput(exec.Command("openssl", "rsa", "-in", file, "-pubout", "-out", pubPath)); err != nil {
			fmt.Println(string(out))
			panic(err)
		}
	}
}

// Sign signs an executable built with ego-go.
func (c *Cli) Sign(filename string) error {
	if filename == "" {
		// When no filename is defined, use an existing enclave.json for reference, or fail otherwise.
		conf, err := c.readConfigJSONtoStruct(defaultConfigFilename)
		if err != nil {
			return err
		}
		return c.signWithJSON(conf)
	}
	if filepath.Ext(filename) == ".json" {
		// If the supplied filename seems to be a JSON file, interpret it as a config file.
		conf, err := c.readConfigJSONtoStruct(filename)
		if err != nil {
			return err
		}
		return c.signWithJSON(conf)
	}
	// If no config exists or no JSON file has been specified, assume the supplied file name is an executable.
	return c.signExecutable(filename)
}
