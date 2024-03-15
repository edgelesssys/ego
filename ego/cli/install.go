// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/klauspost/cpuid/v2"
)

const shellToUse = "bash"

var (
	ErrTargetNotSupported = errors.New("component not found")
	ErrInstallUserQuit    = errors.New("user denied installation")
	ErrExitCodeValue      = errors.New("exit code not 0")
	ErrSysInfoFail        = errors.New("could not determine necessary details about operating system")
)

type installInfoV1 struct {
	Desc     map[string]string
	Commands map[string]map[string][]string
}

type installInfo struct {
	V1 installInfoV1 `json:"1.0"`
}

func (i installInfo) getDescriptions() map[string]string {
	return i.V1.Desc
}

func (i installInfo) getCommands(osDetails string) map[string][]string {
	return i.V1.Commands[osDetails]
}

// Determine relevant information about os and sgx compatibility. Then pass them with osDetails to install
func (c *Cli) Install(ask func(string) bool, component string) error {
	var sgxLevel string
	if !cpuid.CPU.Supports(cpuid.SGX) { // determine support for sgx
		sgxLevel = "nonsgx"
	} else if cpuid.CPU.Supports(cpuid.SGXLC) { // determine support for sgxflc
		sgxLevel = "flc"
	} else {
		sgxLevel = "nonflc"
	}

	url := os.Getenv("EGO_INSTALL_URL")
	if url == "" {
		url = "https://raw.githubusercontent.com/edgelesssys/ego/master/src/install.json"
	} else {
		fmt.Println("Warning: EGO_INSTALL_URL has been set to override installation data url:", url)
	}
	return c.install(ask, sgxLevel, component, url)
}

func (c *Cli) install(ask func(string) bool, sgxLevel string, component string, jsonURL string) error {
	id, versionID, err := c.getOsInfo()
	if err != nil {
		return err
	}
	osDetails := strings.Join([]string{id, versionID, sgxLevel}, "-")

	body, err := httpGet(jsonURL)
	if err != nil {
		return err
	}

	// Parse json data from url into installInfo structure
	var installInfo installInfo
	if err := json.Unmarshal(body, &installInfo); err != nil {
		return err
	}

	// Remove leading white spaces, required to be able to access infoInstall structure
	component = strings.TrimSpace(component)

	osCommands := installInfo.getCommands(osDetails)
	if osCommands == nil { // osDetails not listed in json file
		fmt.Println("No available components to install for your system")
		return nil
	}
	if osCommands[component] == nil { // Print the descriptions of the available components to install. osCommands[component]=nil if the component to install not available for the os. This is also the case if user wants to get description -> in this case no error
		if component != "" {
			fmt.Println("Component not found.")
		}
		descs := installInfo.getDescriptions()
		fmt.Println("Available components that can be installed:")
		for comp := range osCommands {
			fmt.Printf("%-20v %v\n", comp, descs[comp])
		}
		if component == "" {
			return nil
		}
		return ErrTargetNotSupported
	}
	// The system supports the component in the latest version of the json entry. Asks whether the user wants to continue installation
	listOfActions := strings.Join(osCommands[component], "\n")
	if !ask(listOfActions) {
		return ErrInstallUserQuit
	}
	// User wants to continue installation. Extract all commands to install the component from the json entry and run them in the selected shell in a temporary working directory
	dir, err := c.fs.TempDir("", "")
	if err != nil {
		return err
	}
	for _, command := range osCommands[component] {
		fmt.Println("Executing:", command)
		cmd := exec.Command(shellToUse, "-c", command)
		cmd.Dir = dir
		if err := c.runCmd(cmd); err != nil {
			return err
		}
	}
	// Remove the temporary directory
	if err := c.fs.RemoveAll(dir); err != nil {
		return err
	}
	return nil
}

func (c *Cli) runCmd(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := c.runner.Run(cmd); err != nil {
		return err
	}
	if exitcode := c.runner.ExitCode(cmd); exitcode != 0 {
		fmt.Printf("Command exited with Exit Status: %d\n", exitcode)
		return ErrExitCodeValue
	}
	return nil
}

func (c *Cli) getOsInfo() (id, versionID string, err error) {
	const reValue = `"?([^"]+)"?` // optionally quoted
	reID := regexp.MustCompile("^ID=" + reValue + "$")
	reVersionID := regexp.MustCompile("^VERSION_ID=" + reValue + "$")

	const filename = "/etc/os-release"
	file, err := c.fs.Open(filename)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if match := reID.FindStringSubmatch(text); match != nil {
			id = match[1]
		} else if match := reVersionID.FindStringSubmatch(text); match != nil {
			versionID = match[1]
		}
	}

	if id == "" {
		return "", "", errors.New("value 'ID' not found in " + filename)
	}
	if versionID == "" {
		return "", "", errors.New("value 'VersionID' not found in " + filename)
	}
	return id, versionID, nil
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http response has status %v", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
