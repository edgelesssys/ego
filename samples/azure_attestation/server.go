package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/edgelesssys/ego/enclave"
)

// serverAddr is the address of the server
const serverAddr = "0.0.0.0:8080"

// attestationProviderURL is the URL of the attestation provider
const attestationProviderURL = "https://shareduks.uks.attest.azure.net"

func main() {
	// Create a self signed certificate.
	cert, priv := createCertificate()
	fmt.Println("🆗 Generated Certificate.")

	// Cerate an Azure Attestation Token.
	token, err := enclave.CreateAzureAttestationToken(cert, attestationProviderURL)
	if err != nil {
		panic(err)
	}
	fmt.Println("🆗 Created an Microsoft Azure Attestation Token.")

	// Create HTTPS server.
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(token)) })
	http.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("📫 %v sent secret %v\n", r.RemoteAddr, r.URL.Query()["s"])
	})

	tlsCfg := tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{cert},
				PrivateKey:  priv,
			},
		},
	}

	server := http.Server{Addr: serverAddr, TLSConfig: &tlsCfg}
	fmt.Printf("📎 Token now available under https://%s/token\n", serverAddr)
	fmt.Printf("👂 Listening on https://%s/secret for secrets...\n", serverAddr)
	err = server.ListenAndServeTLS("", "")
	fmt.Println(err)
}

func createCertificate() ([]byte, crypto.PrivateKey) {
	template := &x509.Certificate{
		SerialNumber: &big.Int{},
		Subject:      pkix.Name{CommonName: "localhost"},
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     []string{"localhost"},
	}
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	cert, _ := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	return cert, priv
}
