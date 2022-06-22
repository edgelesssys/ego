package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/edgelesssys/ego/enclave"
)

func main() {
	const attestURL = "http://localhost:8080"
	const secureURL = "https://localhost:8081"

	// create client keys
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubKey := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)

	// get server certificate over insecure channel
	serverCert := httpGet(nil, attestURL+"/cert")

	// get the server's report targeted at this client
	clientInfoReport, err := enclave.GetLocalReport(nil, nil)
	if err != nil {
		panic(err)
	}
	serverReport := httpGet(nil, attestURL+"/report", makeArg("target", clientInfoReport))

	// verify server certificate using the server's report
	if err := verifyReport(serverReport, serverCert); err != nil {
		panic(err)
	}

	// request a client certificate from the server
	pubKeyHash := sha256.Sum256(pubKey)
	clientReport, err := enclave.GetLocalReport(pubKeyHash[:], serverReport)
	if err != nil {
		panic(err)
	}
	clientCert := httpGet(nil, attestURL+"/client", makeArg("pubkey", pubKey), makeArg("report", clientReport))

	// create mutual TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{clientCert},
				PrivateKey:  privKey,
			},
		},
		RootCAs: x509.NewCertPool(),
	}
	parsedServerCert, _ := x509.ParseCertificate(serverCert)
	tlsConfig.RootCAs.AddCert(parsedServerCert)

	// use the established secure channel
	resp := httpGet(tlsConfig, secureURL+"/ping")
	fmt.Printf("server responded: %s\n", resp)
}

func verifyReport(reportBytes []byte, cert []byte) error {
	report, err := enclave.VerifyLocalReport(reportBytes)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(cert)
	if !bytes.Equal(report.Data[:len(hash)], hash[:]) {
		return errors.New("report data doesn't match the server certificate's hash")
	}

	// We expect the other enclave to be signed with the same key.

	selfReport, err := enclave.GetSelfReport()
	if err != nil {
		return err
	}

	if !bytes.Equal(report.SignerID, selfReport.SignerID) {
		return errors.New("invalid signer")
	}
	if binary.LittleEndian.Uint16(report.ProductID) != 2 {
		return errors.New("invalid product")
	}
	if report.SecurityVersion < 1 {
		return errors.New("invalid security version")
	}
	if report.Debug && !selfReport.Debug {
		return errors.New("other party is a debug enclave")
	}

	return nil
}

func httpGet(tlsConfig *tls.Config, url string, args ...string) []byte {
	client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	fmt.Println("GET " + url)
	if len(args) > 0 {
		url += "?" + strings.Join(args, "&")
	}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(os.Stdout, resp.Body)
		panic(resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return body
}

func makeArg(key string, value []byte) string {
	return key + "=" + base64.URLEncoding.EncodeToString(value)
}
