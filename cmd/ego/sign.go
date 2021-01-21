package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const defaultConfigFilename = "enclave.json"
const defaultKeyFilename = "private.pem"

type config struct {
	Exe             string `json:"exe"`
	Key             string `json:"key"`
	Debug           bool   `json:"debug"`
	Heapsize        int    `json:"heapsize"`
	ProductID       int    `json:"productID"`
	SecurityVersion int    `json:"securityVersion"`
}

// Validate Exe, Key, Heapsize
func (c *config) validate() {
	if c.Heapsize == 0 {
		panic(fmt.Errorf("heapsize not set in enclave.json"))
	}
	if c.Exe == "" {
		panic(fmt.Errorf("exe not set in enclave.json"))
	}
	if c.Key == "" {
		panic(fmt.Errorf("key not set in enclave.json"))
	}
}

func signWithJSON(conf *config) {
	//write temp .conf file
	cProduct := "ProductID=" + strconv.Itoa(conf.ProductID) + "\n"
	cSecurityVersion := "SecurityVersion=" + strconv.Itoa(conf.SecurityVersion) + "\n"

	var cDebug string
	if conf.Debug {
		cDebug = "Debug=1\n"
	} else {
		cDebug = "Debug=0\n"
	}

	// calculate number of pages: Heapsize[MiB], pageSize is 4096B
	heapPages := conf.Heapsize * 1024 * 1024 / 4096
	cNumHeapPages := "NumHeapPages=" + strconv.Itoa(heapPages) + "\n"

	cStackPages := "NumStackPages=1024\n"
	cNumTCS := "NumTCS=32\n"

	file, err := ioutil.TempFile("", "")
	defer os.Remove(file.Name())
	_, err = file.Write([]byte(cProduct + cSecurityVersion + cDebug + cNumHeapPages + cStackPages + cNumTCS))

	if err != nil {
		panic(err)
	}

	if err := file.Close(); err != nil {
		panic(err)
	}

	enclavePath := filepath.Join(egoPath, "share", "ego-enclave")
	cmd := exec.Command("ego-oesign", "sign", "-e", enclavePath, "-c", file.Name(), "-k", conf.Key, "--payload", conf.Exe)
	runAndExit(cmd)
}

func signExecutable(path string) {
	c, err := readJSONtoStruct(defaultConfigFilename)

	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else if c.Exe == path {
		signWithJSON(c)
	} else {
		panic(fmt.Errorf("Provided path to executable does not match the one in enclave.json"))
	}

	//sane default values
	conf := config{
		Exe:             path,
		Key:             defaultKeyFilename,
		Debug:           true,
		Heapsize:        512, //[MB]
		ProductID:       1,
		SecurityVersion: 1,
	}

	jsonData, err := json.MarshalIndent(conf, "", " ")
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(defaultConfigFilename, jsonData, 0644); err != nil {
		panic(err)
	}

	if _, err := os.Stat(defaultKeyFilename); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		// SGX requires the RSA exponent to be 3. Go's API does not support this.
		if err := exec.Command("openssl", "genrsa", "-out", defaultKeyFilename, "-3", "3072").Run(); err != nil {
			panic(err)
		}
	}

	c, err = readJSONtoStruct(defaultConfigFilename)
	if err != nil {
		panic(err)
	}
	signWithJSON(c)
}

// Reads the provided File and turns it into a struct
// after some basic sanity check are performed it is returned
// err != nil indicates that the file could not be read or the
// JSON could not be unmarshalled
func readJSONtoStruct(path string) (*config, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var conf config
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	conf.validate()
	return &conf, nil
}

func sign(filename string) {
	if filename == "" {
		c, err := readJSONtoStruct(defaultConfigFilename)
		if err != nil {
			panic(err)
		}
		signWithJSON(c)
	}
	if filepath.Ext(filename) == ".json" {
		c, err := readJSONtoStruct(filename)
		if err != nil {
			panic(err)
		}
		signWithJSON(c)
	}
	signExecutable(filename)
}
