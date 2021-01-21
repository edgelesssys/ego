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
	if len(os.Args) < 2 {
		help("")
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "sign":
		var filename string
		if len(args) > 0 {
			filename = args[0]
		}
		sign(filename)
		return
	case "run":
		if len(args) > 0 {
			run(args[0], args[1:])
			return
		}
	case "env":
		if len(args) > 0 {
			env(args[0], args[1:])
			return
		}
	case "help":
		if len(args) == 1 {
			help(args[0])
			return
		}
	}

	help(cmd)
}

func help(cmd string) {
	me := os.Args[0]
	var s string

	switch cmd {
	case "sign":
		s = `sign [executable|json]

Sign an executable built with ego-go. Executables must be signed before they can be run in an enclave.
There are 3 different ways of signing an executable:

1. ego sign 
Searches in the current directory for enclave.json and signs the therein provided executable.

2. ego sign <executable>
Searches in the current directory for enclave.json.
If no such file is found, create one with default parameters for the executable.
If a file already exists, check if the executable matches the parameter. 
Use enclave.json to sign the executable.

3. ego sign <json>
Reads json and signs the therein provided executable.`

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
