package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/edgelesssys/ego/enclave"
)

func main() {
	serverAddr := "0.0.0.0:8080"
	// Create certificate and a report that includes the certificate's hash.
	cert, priv := createCertificate()
	fmt.Println("🆗 Generated Certificate.")
	hash := sha256.Sum256(cert)
	report, err := enclave.GetRemoteReport(hash[:])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("🆗 Got a report from the enclave.")

	// Create Attestation Request
	token := azureAttestation(report, cert, "https://shareduks.uks.attest.azure.net")

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

type attestOERequest struct {
	Report      string `json:"report"`
	RuntimeData rtdata `json:"runtimeData"`
}

type rtdata struct {
	Data     string `json:"data"`
	DataType string `json:"dataType"`
}

type attestationResponse struct {
	Token string `json:"token"`
}

func azureAttestation(report, data []byte, uri string) string {
	rtd := rtdata{Data: base64.RawURLEncoding.EncodeToString(data), DataType: "Binary"}
	attReq := attestOERequest{Report: base64.RawURLEncoding.EncodeToString(report), RuntimeData: rtd}
	// Build URL.
	azUrl, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}
	azPath, err := url.Parse("/attest/OpenEnclave?api-version=2020-10-01")
	if err != nil {
		panic(err)
	}
	azUrl = azUrl.ResolveReference(azPath)

	// Send POST request with JSON body.
	jsonVal, err := json.Marshal(attReq)
	if err != nil {
		panic(err)
	}
	resp, err := http.Post(azUrl.String(), "application/json", bytes.NewReader(jsonVal))
	if err != nil {
		panic(err)
	}
	fmt.Println("📨 Sent Attestation Request which contains report and cerificate.")

	// Check and parse response.
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		s, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(s))
		panic(resp.Status)
	}
	body := new(attestationResponse)
	if err := json.NewDecoder(resp.Body).Decode(body); err != nil {
		panic(err)
	}
	fmt.Println("📨 Recived Attestation Response with token.")
	return body.Token
}
