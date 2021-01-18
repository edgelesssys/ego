package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type config struct {
	Exe             string `json:"exe"`
	Key             string `json:"key"`
	Debug           bool   `json:"debug"`
	Heapsize        int    `json:"heapsize"`
	ProductID       int    `json:"productID"`
	SecurityVersion int    `json:"securityVersion"`
}

// Validate Exe, Key, Heapsize
func validateJSON(c *config) {
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

func signWithJSON(path string) {
	conf, err := readJSONtoStruct(path)
	if err != nil {
		panic(err)
	}

	//write temp .conf file
	cProduct := "ProductID=" + strconv.Itoa(conf.ProductID) + "\n"
	cSecurityVersion := "SecurityVersion=" + strconv.Itoa(conf.SecurityVersion) + "\n"

	var cDebug string
	if conf.Debug {
		cDebug = "Debug=1\n"
	} else {
		cDebug = "Debug=0\n"
	}

	// calculate number of pages: Heapsize[MiB], pageSize is 4096KiB
	heapPages := int(math.Ceil(float64(conf.Heapsize*1024) / float64(4096)))
	cNumHeapPages := "NumHeapPages=" + strconv.Itoa(heapPages) + "\n"

	cStackPages := "NumStackPages=1024\n"
	cNumTCS := "NumTCS=32\n"

	err = ioutil.WriteFile("config.conf.tmp", []byte(cProduct+cSecurityVersion+cDebug+cNumHeapPages+cStackPages+cNumTCS), 0644)
	if err != nil {
		panic(err)
	}

	enclavePath := filepath.Join(egoPath, "share", "ego-enclave")
	cmd := exec.Command("ego-oesign", "sign", "-e", enclavePath, "-c", "config.conf.tmp", "-k", conf.Key, "--payload", conf.Exe)
	runAndExit(cmd)
}

func signExecutable(path string) {
	c, err := readJSONtoStruct("enclave.json")

	if c != nil && c.Exe == path {
		signWithJSON("enclave.json")
	} else if c != nil && c.Exe != path {
		panic(fmt.Errorf("Provided path to executable does not match the one in enclave.json"))
	}

	//sane default values
	conf := config{
		Exe:             path,
		Key:             "private.pem",
		Debug:           true,
		Heapsize:        512, //[MB]
		ProductID:       1,
		SecurityVersion: 1,
	}

	jsonData, err := json.MarshalIndent(conf, "", " ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("enclave.json", jsonData, 0644)
	if err != nil {
		panic(err)
	}

	_, err = os.Stat("private.pem")
	if err != nil {
		// SGX requires the RSA exponent to be 3. Go's API does not support this.
		if err := exec.Command("openssl", "genrsa", "-out", "private.pem", "-3", "3072").Run(); err != nil {
			panic(err)
		}
	}
	signWithJSON("enclave.json")
}

// Reads the provided File and turns it into a struct
// after some basic sanity check are performed it is returned
// err != nil indicates that the file could not be read
func readJSONtoStruct(path string) (*config, error) {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var conf config
	err = json.Unmarshal(data, &conf)
	if err != nil {
		panic(err)
	}
	validateJSON(&conf)
	return &conf, nil
}

func sign(args []string) {
	if len(args) < 3 {
		signWithJSON("enclave.json")
	}
	if strings.HasSuffix(args[2], "enclave.json") {
		signWithJSON(args[2])
	}
	signExecutable(args[2])
}
