package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var egoPath = func() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(filepath.Dir(exe)) // parent dir of dir of exe
}()

func main() {
	if len(os.Args) < 3 {
		help("")
		return
	}

	switch os.Args[1] {
	case "sign":
		sign(os.Args[2])
	case "run":
		run(os.Args[2], os.Args[3:])
	case "env":
		env(os.Args[2], os.Args[3:])
	case "help":
		help(os.Args[2])
	default:
		help("")
	}
}

func help(cmd string) {
	me := os.Args[0]
	var s string

	switch cmd {
	case "sign":
		s = `sign <executable>

Sign an executable built with ego-go. Executables must be signed before they can be run in an enclave.`
	case "run":
		s = `run <executable> [args...]

Run a signed executable in an enclave. You can pass arbitrary arguments to the enclave.

Environment variables are only readable from within the enclave if they start with "EDG_".

Set OE_SIMULATION=1 to run in simulation mode.`
	case "env":
		s = `env ...

Run a command within the ego environment. For example, run
` + me + ` env make
to build a Go project that uses a Makefile.`
	default:
		s = `<command> [arguments]

Commands:
  sign   Sign an executable built with ego-go.
  run    Run a signed executable.
  env    Run a command in the ego environment.

Use "` + me + ` help <command>" for more information about a command.`
	}

	fmt.Println("Usage: " + me + " " + s)
}

func sign(filename string) {
	// SGX requires the RSA exponent to be 3. Go's API does not support this.
	if err := exec.Command("openssl", "genrsa", "-out", "private.pem", "-3", "3072").Run(); err != nil {
		panic(err)
	}

	enclavePath := filepath.Join(egoPath, "share", "ego-enclave")
	confPath := filepath.Join(egoPath, "share", "enclave.conf")
	cmd := exec.Command("ego-oesign", "sign", "-e", enclavePath, "-c", confPath, "-k", "private.pem", "--payload", filename)
	runAndExit(cmd)
}

func run(filename string, args []string) {
	enclaves := filepath.Join(egoPath, "share", "ego-enclave") + ":" + filename
	args = append([]string{enclaves}, args...)
	cmd := exec.Command("ego-host", args...)
	runAndExit(cmd)
}

func env(filename string, args []string) {
	if filename == "go" {
		// "ego env go" should resolve to our Go compiler
		filename = filepath.Join(egoPath, "go", "bin", "go")
	}
	cmd := exec.Command(filename, args...)
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=1",
		"PATH="+filepath.Join(egoPath, "go", "bin")+":"+os.Getenv("PATH"),
		"GOROOT="+filepath.Join(egoPath, "go"))
	runAndExit(cmd)
}

func runAndExit(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			panic(err)
		}
	}
	os.Exit(cmd.ProcessState.ExitCode())
}
