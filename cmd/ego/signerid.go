package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

type eradump struct {
	SecurityVersion int    `json:"SecurityVersion"`
	ProductID       int    `json:"ProductID"`
	UniqueID        string `json:"UniqueID"`
	SignerID        string `json:"SignerID"`
}

func signeridByKey(path string) {
	out, err := exec.Command("ego-oesign", "signerid", "-k", path).Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

func readEradumpJSONtoStruct(path string) *eradump {
	data, err := exec.Command("ego-oesign", "eradump", "-e", path).Output()

	if err != nil {
		panic(err)
	}

	var dump eradump
	if err := json.Unmarshal(data, &dump); err != nil {
		panic(err)
	}
	return &dump
}

func signeridByExecutable(path string) {
	dump := readEradumpJSONtoStruct(path)
	fmt.Println(dump.SignerID)
}

func uniqueid(path string) {
	dump := readEradumpJSONtoStruct(path)
	fmt.Println(dump.UniqueID)
}

func signerid(path string) {
	if filepath.Ext(path) == ".pem" {
		signeridByKey(path)
	} else {
		signeridByExecutable(path)
	}
}
