// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type installerRunner struct {
	run []*exec.Cmd
}

func (i *installerRunner) Run(cmd *exec.Cmd) error {
	i.run = append(i.run, cmd)
	return nil
}

func (*installerRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	panic(cmd.Path)
}

func (*installerRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	panic(cmd.Path)
}

func (*installerRunner) ExitCode(cmd *exec.Cmd) int {
	return 0
}

func (i *installerRunner) ClearRun() {
	i.run = make([]*exec.Cmd, 0)
}

const ubuntu1804 = "ID=ubuntu\nVERSION_ID=18.04"
const ubuntu2004 = "ID=ubuntu\nVERSION_ID=20.04"

var jsonData = `
{
	"1.0": {
		"desc": {
			"sgx-driver": "The SGX driver. Required on systems with a Linux kernel version prior to 5.11.",
			"libsgx-launch": "SGX Launch package required for non-FLC SGX systems.",
			"az-dcap-client": "The Microsoft Azure DCAP client. Required for remote attestation with an Azure VM."
        },
        "commands": {
			"ubuntu-18.04-flc": {
				"sgx-driver": [
					"! test -e /dev/sgx && ! test -e /dev/isgx",
					"wget https://download.01.org/intel-sgx/sgx-linux/2.12/distro/ubuntu18.04-server/sgx_linux_x64_driver_1.36.2.bin",
					"echo '7b415b8a4367b3deb044ee7d94fffb7922df17a3da2a238da3f8db9391ecb650  sgx_linux_x64_driver_1.36.2.bin' | sha256sum -c",
					"chmod +x sgx_linux_x64_driver_1.36.2.bin",
					"./sgx_linux_x64_driver_1.36.2.bin"
				],
				"az-dcap-client": [
					"wget -q https://packages.microsoft.com/keys/microsoft.asc",
					"echo '2cfd20a306b2fa5e25522d78f2ef50a1f429d35fd30bd983e2ebffc2b80944fa  microsoft.asc' | sha256sum -c",
					"apt-key add microsoft.asc",
					"add-apt-repository 'deb [arch=amd64] https://packages.microsoft.com/ubuntu/18.04/prod bionic main'",
					"apt install -y az-dcap-client"
				]
            },
            "ubuntu-18.04-nonflc": {
                "sgx-driver": [
                    "! test -e /dev/sgx && ! test -e /dev/isgx",
                    "wget https://download.01.org/intel-sgx/sgx-linux/2.13.3/distro/ubuntu18.04-server/sgx_linux_x64_driver_2.11.0_2d2b795.bin",
					"echo 'd8080eb99a7916ff957c3a187bb46c678e84c4948a671863ac573d4572dd374e  sgx_linux_x64_driver_2.11.0_2d2b795.bin' | sha256sum -c",
                    "chmod +x sgx_linux_x64_driver_2.11.0_2d2b795.bin",
                    "./sgx_linux_x64_driver_2.11.0_2d2b795.bin"
                ],
                "libsgx-launch": [
                    "wget -q https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key",
					"echo '809cb39c4089843f923143834550b185f2e6aa91373a05c8ec44026532dab37c  intel-sgx-deb.key' | sha256sum -c",
					"apt-key add intel-sgx-deb.key",
                    "add-apt-repository 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu bionic main'",
                    "apt install -y libsgx-launch"
                ]
            },
            "ubuntu-20.04-flc": {
                "sgx-driver": [
                    "! test -e /dev/sgx && ! test -e /dev/isgx",
                    "wget https://download.01.org/intel-sgx/sgx-linux/2.12/distro/ubuntu20.04-server/sgx_linux_x64_driver_1.36.2.bin",
					"echo 'e3e027b16dfe980859035a9437b198d21ca9d71825d2144756ac3eac759c489f sgx_linux_x64_driver_1.36.2.bin' | sha256sum -c",
                    "chmod +x sgx_linux_x64_driver_1.36.2.bin",
                    "./sgx_linux_x64_driver_1.36.2.bin"
                ],
                "az-dcap-client": [
                    "wget -q https://packages.microsoft.com/keys/microsoft.asc",
					"echo '2cfd20a306b2fa5e25522d78f2ef50a1f429d35fd30bd983e2ebffc2b80944fa  microsoft.asc' | sha256sum -c",
					"apt-key add microsoft.asc",
                    "add-apt-repository 'deb [arch=amd64] https://packages.microsoft.com/ubuntu/20.04/prod focal main'",
                    "apt install -y az-dcap-client"
                ]
            },
            "ubuntu-20.04-nonflc": {
                "sgx-driver": [
                    "! test -e /dev/sgx && ! test -e /dev/isgx",
                    "wget https://download.01.org/intel-sgx/sgx-linux/2.13.3/distro/ubuntu20.04-server/sgx_linux_x64_driver_2.11.0_2d2b795.bin",
					"echo '8727e5441e6cccc3a77b72c747cc2eeda59ed20ef8502fa73d4865e50fc4d6fd  sgx_linux_x64_driver_2.11.0_2d2b795.bin' | sha256sum -c",
					"chmod +x sgx_linux_x64_driver_2.11.0_2d2b795.bin",
                    "./sgx_linux_x64_driver_2.11.0_2d2b795.bin"
                ],
                "libsgx-launch": [
                    "wget -q https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key",
					"echo '809cb39c4089843f923143834550b185f2e6aa91373a05c8ec44026532dab37c  intel-sgx-deb.key' | sha256sum -c",
					"apt-key add intel-sgx-deb.key",
                    "add-apt-repository 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu focal main'",
                    "apt install -y libsgx-launch"
                ]
            }
        }
    }
}
`

// Test whether getOsInfo can correctly determine details from os-release
func TestInstallGetOsInfo(t *testing.T) {
	assert := assert.New(t)
	runner := installerRunner{}
	fs := afero.NewMemMapFs()

	cli := NewCli(&runner, fs)

	cli.fs.WriteFile("/etc/os-release", []byte("ID=\"ubuntu\"\nsome other infos\nVERSION_ID=\"20.04\""), 0)
	id, versionID, err := cli.getOsInfo()
	assert.Equal("ubuntu", id)
	assert.Equal("20.04", versionID)
	assert.Equal(nil, err)

	cli.fs.WriteFile("/etc/os-release", []byte("ID=foo\nVERSION_ID=bar"), 0)
	id, versionID, err = cli.getOsInfo()
	assert.Equal("foo", id)
	assert.Equal("bar", versionID)
	assert.Equal(nil, err)

	cli.fs.WriteFile("/etc/os-release", []byte("VERSION_ID=20.04"), 0)
	id, versionID, err = cli.getOsInfo()
	assert.Equal("", id)
	assert.Equal("", versionID)
	assert.NotEqual(nil, err)

	cli.fs.WriteFile("/etc/os-release", []byte("IID=ubuntu\nVERSION_ID=20.04"), 0)
	id, versionID, err = cli.getOsInfo()
	assert.Equal("", id)
	assert.Equal("", versionID)
	assert.NotEqual(nil, err)
}

// Run tests that should all pass the installation
func TestInstallValidTests(t *testing.T) {
	assert := assert.New(t)

	runner := installerRunner{}
	cli := NewCli(&runner, afero.NewMemMapFs())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, jsonData)
	}))
	askTrue := func(string) bool { return true }

	validNonflcTests := make(map[string][]string)

	validNonflcTests[ubuntu1804] = []string{""}
	validNonflcTests[ubuntu1804] = []string{"sgx-driver", "libsgx-launch"}
	validNonflcTests[ubuntu2004] = []string{" sgx-driver", "  libsgx-launch"}

	fmt.Println("Valid nonflc tests:")
	fmt.Println("------------------------------------------------------------------------------------")
	for osReleaseData, testComponents := range validNonflcTests {
		cli.fs.WriteFile("/etc/os-release", []byte(osReleaseData), 0600)
		for _, component := range testComponents {
			fmt.Print("\nStarting installation of \"", component, "\"\n")
			assert.Equal(nil, cli.install(askTrue, "nonflc", component, server.URL))
			switch strings.TrimSpace(component) {
			case "sgx-driver":
				assert.Len(runner.run, 5)
				assert.Equal(runner.run[0].Args[0], shellToUse)
			case "libsgx-launch":
				assert.Len(runner.run, 5)
				assert.Equal(runner.run[0].Args[0], shellToUse)
			case "":
				assert.Len(runner.run, 0)
			}
			runner.ClearRun()
			fmt.Println("------------------------------------------------------------------------------------")
		}
	}

	validFlcTests := make(map[string][]string)

	validFlcTests[ubuntu1804] = []string{"sgx-driver", "az-dcap-client", " sgx-driver", "  az-dcap-client"}

	fmt.Println("Valid flc tests:")
	fmt.Println("------------------------------------------------------------------------------------")
	for osReleaseData, testComponents := range validFlcTests {
		cli.fs.WriteFile("/etc/os-release", []byte(osReleaseData), 0600)
		for _, component := range testComponents {
			fmt.Print("\nStarting installation of \"", component, "\"\n")
			assert.Equal(nil, cli.install(askTrue, "flc", component, server.URL))
			switch strings.TrimSpace(component) {
			case "sgx-driver":
				assert.Len(runner.run, 5)
				assert.Equal(runner.run[0].Args[0], shellToUse)
			case "az-dcap-client":
				assert.Len(runner.run, 5)
				assert.Equal(runner.run[0].Args[0], shellToUse)
			case "":
				assert.Len(runner.run, 0)
			}
			runner.ClearRun()
			fmt.Println("------------------------------------------------------------------------------------")
		}
	}

}

// Run tests that should all fail the installation process
func TestInstallNotValidTests(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	runner := installerRunner{}
	cli := NewCli(&runner, afero.NewMemMapFs())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, jsonData)
	}))

	askTrue := func(string) bool { return true }
	fmt.Println("\n\nUnvalid tests:")
	fmt.Println("------------------------------------------------------------------------------------")

	unvalidTests := make(map[string][]string)
	unvalidTests["foo"] = []string{"sgx-driver", "bar"}
	unvalidTests[""] = []string{""}
	unvalidTests[ubuntu2004] = []string{"az-dcap-client", "echo abc", "?libsgx-launch", "! libsgx-launch", "|libsgx-launch", "  . libsgx-launch"}

	for osReleaseData, testComponents := range unvalidTests {
		cli.fs.WriteFile("/etc/os-release", []byte(osReleaseData), 0600)
		for _, component := range testComponents {
			fmt.Print("\nStarting installation of \"", component, "\"\n")
			assert.NotEqual(nil, cli.install(askTrue, "nonflc", component, server.URL))
			require.Len(runner.run, 0)
			fmt.Println("------------------------------------------------------------------------------------")
		}
		runner.ClearRun()
	}
}

// Tests that all commands in this specific installation match the expected ones
func TestExactCommand(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	runner := installerRunner{}
	cli := NewCli(&runner, afero.NewMemMapFs())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, jsonData)
	}))

	askTrue := func(string) bool { return true }

	fmt.Println("\n\nExact Command Test")
	fmt.Println("------------------------------------------------------------------------------------")

	cli.fs.WriteFile("/etc/os-release", []byte(ubuntu1804), 0600)
	assert.Equal(nil, cli.install(askTrue, "flc", "sgx-driver", server.URL))
	cmds := runner.run
	fmt.Println(cmds[0].Dir)
	require.Len(runner.run, 5)
	assert.Equal(cmds[0].Args, []string{shellToUse, "-c", "! test -e /dev/sgx && ! test -e /dev/isgx"})
	assert.Equal(cmds[1].Args, []string{shellToUse, "-c", "wget https://download.01.org/intel-sgx/sgx-linux/2.12/distro/ubuntu18.04-server/sgx_linux_x64_driver_1.36.2.bin"})
	assert.Equal(cmds[2].Args, []string{shellToUse, "-c", "echo '7b415b8a4367b3deb044ee7d94fffb7922df17a3da2a238da3f8db9391ecb650  sgx_linux_x64_driver_1.36.2.bin' | sha256sum -c"})
	assert.Equal(cmds[3].Args, []string{shellToUse, "-c", "chmod +x sgx_linux_x64_driver_1.36.2.bin"})
	assert.Equal(cmds[4].Args, []string{shellToUse, "-c", "./sgx_linux_x64_driver_1.36.2.bin"})
	runner.ClearRun()
	fmt.Println("------------------------------------------------------------------------------------")
}

func TestInstallErrorCheck(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	runner := installerRunner{}
	cli := NewCli(&runner, afero.NewMemMapFs())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, jsonData)
	}))

	askTrue := func(string) bool { return true }
	askFalse := func(string) bool { return false }

	fmt.Println("\n\nTest Install Errors")

	// os-release file does not contain necessary information, but no error
	cli.fs.WriteFile("/etc/os-release", []byte(""), 0600)
	assert.NotEqual(nil, cli.install(askTrue, "flc", "sgx-driver", server.URL))
	runner.ClearRun()
	fmt.Println("------------------------------------------------------------------------------------")

	// os "foo" does not exist in json file
	cli.fs.WriteFile("/etc/os-release", []byte("ID=foo\nVERSION_ID=bar"), 0600)
	assert.Equal(nil, cli.install(askTrue, "flc", "sgx-driver", server.URL))
	runner.ClearRun()
	fmt.Println("------------------------------------------------------------------------------------")

	// no available components to install for nonsgx
	cli.fs.WriteFile("/etc/os-release", []byte(ubuntu2004), 0600)
	assert.Equal(nil, cli.install(askTrue, "nonsgx", "sgx-driver", server.URL))
	runner.ClearRun()
	fmt.Println("------------------------------------------------------------------------------------")

	// component "foo" does not exist in json file
	cli.fs.WriteFile("/etc/os-release", []byte(ubuntu2004), 0600)
	assert.Equal(ErrTargetNotSupported, cli.install(askTrue, "nonflc", "foo", server.URL))
	runner.ClearRun()
	fmt.Println("------------------------------------------------------------------------------------")

	// askFalse: user does not want to continue installation, so installation stops without error
	cli.fs.WriteFile("/etc/os-release", []byte(ubuntu2004), 0600)
	assert.Equal(ErrInstallUserQuit, cli.install(askFalse, "nonflc", "sgx-driver", server.URL))
	require.Len(runner.run, 0)
	runner.ClearRun()
	fmt.Println("------------------------------------------------------------------------------------")
}

// Checks whether the exit codes of the runCmd function are correct
func TestRunCmd(t *testing.T) {
	assert := assert.New(t)

	runner := installCmdRunner{}
	cli := NewCli(&runner, afero.NewMemMapFs())

	fmt.Println("\n\nTest runCmd() Method")
	fmt.Println("------------------------------------------------------------------------------------")
	fmt.Println("Executing: echo abc")
	assert.Equal(nil, cli.runCmd(exec.Command(shellToUse, "-c", "echo abc")))
	fmt.Println("Executing: abc")
	assert.NotEqual(nil, cli.runCmd(exec.Command(shellToUse, "-c", "abc")))
	fmt.Println("Executing: ! test -e /")
	assert.NotEqual(nil, cli.runCmd(exec.Command(shellToUse, "-c", "! test -e /")))
	fmt.Println("------------------------------------------------------------------------------------")
}

type installCmdRunner struct{}

func (installCmdRunner) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (installCmdRunner) Output(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

func (installCmdRunner) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

func (installCmdRunner) ExitCode(cmd *exec.Cmd) int {
	return cmd.ProcessState.ExitCode()
}
