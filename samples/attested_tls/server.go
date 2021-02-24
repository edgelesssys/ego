package main

import (
	"fmt"
	"net/http"

	"github.com/edgelesssys/ego/enclave"
)

func main() {
	// Create a TLS config with a self-signed certificate and an embedded report.
	tlsCfg, err := enclave.CreateAttestationServerTLSConfig()
	if err != nil {
		panic(err)
	}

	// Create HTTPS server.

	http.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%v sent secret %v\n", r.RemoteAddr, r.URL.Query()["s"])
	})

	server := http.Server{Addr: "0.0.0.0:8080", TLSConfig: tlsCfg}

	fmt.Println("listening ...")
	err = server.ListenAndServeTLS("", "")
	fmt.Println(err)
}
