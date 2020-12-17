package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	egoPath, _ := os.Executable()
	egoPath = filepath.Dir(filepath.Dir(egoPath))

	switch os.Args[1] {
	case "sign":
		sign(os.Args[2], egoPath)
	// case "run":
	// 	enclaves := filepath.Join(egoPath, "share", "ego-enclave") + ":" + os.Args[2]
	// 	args := append([]string{enclaves}, os.Args[3:]...)
	// 	if err := exec.Command("ego-host", args...).Start(); err != nil {
	// 		panic(err)
	// 	}
	case "path":
		fmt.Print(filepath.Join(egoPath, "go", "bin") + ":" + os.Getenv("PATH"))
	case "root":
		fmt.Print(filepath.Join(egoPath, "go"))
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: " + os.Args[0] + ` <command>
  sign <executable>
  path
  root`)
}

func sign(filename, egoPath string) {
	// key, _ := rsa.GenerateKey(rand.Reader, 3072)
	// keyRaw := x509.MarshalPKCS1PrivateKey(key)
	// keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyRaw})
	// ioutil.WriteFile("private.pem", keyPem, 0600)
	exec.Command("openssl", "genrsa", "-out", "private.pem", "-3", "3072").Run()

	enclavePath := filepath.Join(egoPath, "share", "ego-enclave")
	confPath := filepath.Join(egoPath, "share", "enclave.conf")
	cmd := exec.Command("ego-sign", "sign", "-e", enclavePath, "-c", confPath, "-k", "private.pem", "--payload", filename)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			panic(err)
		}
	}
	fmt.Println(string(out))
}
