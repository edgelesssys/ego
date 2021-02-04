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
const defaultPrivKeyFilename = "private.pem"
const defaultPubKeyFilename = "public.pem"

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
		panic(fmt.Errorf("heapsize not set in config file"))
	}
	if c.Exe == "" {
		panic(fmt.Errorf("exe not set in config file"))
	}
	if c.Key == "" {
		panic(fmt.Errorf("key not set in config file"))
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
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())

	_, err = file.Write([]byte(cProduct + cSecurityVersion + cDebug + cNumHeapPages + cStackPages + cNumTCS))
	if err != nil {
		panic(err)
	}

	if err := file.Close(); err != nil {
		panic(err)
	}

	//create public and private key if private key does not exits
	createDefaultKeypair(conf.Key)

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
		Key:             defaultPrivKeyFilename,
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

	signWithJSON(&conf)
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

// Creates a public/secret keypair if the provided secret key does not exists
func createDefaultKeypair(file string) {
	if _, err := os.Stat(file); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		fmt.Println("Generating new " + file)
		// SGX requires the RSA exponent to be 3. Go's API does not support this.
		if err := exec.Command("openssl", "genrsa", "-out", file, "-3", "3072").Run(); err != nil {
			panic(err)
		}
		pubPath := filepath.Join(filepath.Dir(file), defaultPubKeyFilename)
		if err := exec.Command("openssl", "rsa", "-in", file, "-pubout", "-out", pubPath).Run(); err != nil {
			panic(err)
		}
	}
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
