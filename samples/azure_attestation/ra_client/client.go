package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/edgelesssys/ego/attestation"
)

// attestationProviderURL is the URL of the attestation provider
const attestationProviderURL = "https://shareduks.uks.attest.azure.net"

func main() {
	signer := flag.String("s", "", "signer ID")
	serverAddr := flag.String("a", "localhost:8080", "server address")
	flag.Parse()

	// Ensure signerID is passed.
	if len(*signer) == 0 {
		flag.Usage()
		return
	}

	// Get attestation token from server. Skip TLS certificate verification because
	// the certificate is self-signed and we will verify it using the token instead.
	serverUrl := "https://" + *serverAddr
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	tokenBytes := httpGet(tlsConfig, serverUrl+"/token")
	fmt.Printf("🆗 Loaded server attestation token from %s.\n", serverUrl+"/token")

	report, err := attestation.VerifyAzureAttestationToken(string(tokenBytes), attestationProviderURL)
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ Azure Attestation Token verified.")

	// Verify the report. ProductID, SecurityVersion and Debug were defined in
	// the enclave.json, and included in the servers binary.
	if err := verifyReportValues(report, *signer); err != nil {
		panic(err)
	}

	// Get certificate from the report.
	certBytes := report.Data
	fmt.Println("🆗 Server certificate extracted from token.")

	// Create a TLS config that uses the server certificate as root
	// CA so that future connections to the server can be verified.
	cert, _ := x509.ParseCertificate(certBytes)
	tlsConfig = &tls.Config{RootCAs: x509.NewCertPool(), ServerName: "localhost"}
	tlsConfig.RootCAs.AddCert(cert)

	httpGet(tlsConfig, serverUrl+"/secret?s=thisIsTheSecret")
	fmt.Println("🔒 Sent secret over attested TLS channel.")
}

// verifyReportValues compares the report values with that were defined in the
// enclave.json and that were included into the binary of the server during build.
func verifyReportValues(report attestation.Report, signer string) error {
	if !reflect.DeepEqual(report.SignerID, []byte(signer)) {
		return errors.New("token does not contain the right signer id")
	}
	fmt.Println("✅ SignerID of the report equals the SignerID you passed to the client.")

	productID := make([]byte, 16)
	binary.LittleEndian.PutUint16(productID, uint16(1234))
	if !reflect.DeepEqual(report.ProductID, productID) {
		return errors.New("token contains invalid product id")
	}
	fmt.Println("✅ ProductID verified.")

	if report.SecurityVersion < 2 {
		return errors.New("token contains invalid security version number")
	}
	fmt.Println("✅ SecurityVersion verified.")
	return nil
}

func httpGet(tlsConfig *tls.Config, url string) []byte {
	client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return body
}
