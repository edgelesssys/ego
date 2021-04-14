package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// issAddr is the URL of the attestation provider
const issAddr = "https://shareduks.uks.attest.azure.net"

// privateClaims are some of the private claims of an Azure Attestation token
type privateClaims struct {
	SignerID        string `json:"x-ms-sgx-mrsigner"`
	ProductID       uint   `json:"x-ms-sgx-product-id"`
	SecurityVersion uint   `json:"x-ms-sgx-svn"`
	EnclaveHeldData string `json:"x-ms-sgx-ehd"`
}

func main() {
	signer := flag.String("s", "", "signer ID")
	serverAddr := flag.String("a", "localhost:8080", "server address")
	flag.Parse()

	// Ensure signer ID is passed.
	if len(*signer) == 0 {
		flag.Usage()
		return
	}

	// Get server token. Skip TLS certificate verification because
	// the certificate is self-signed and we will verify it using the token instead.
	serverUrl := "https://" + *serverAddr
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	tokenBytes := httpGet(tlsConfig, serverUrl+"/token")
	fmt.Printf("🆗 Loaded server attestation token from %s.\n", serverUrl+"/token")

	// Get cerificate for token verification from attestation provider.
	tlsConfig = &tls.Config{}
	jwkSetBytes := httpGet(tlsConfig, issAddr+"/certs")
	keySet := mustParseKeySet(jwkSetBytes)
	fmt.Printf("🆗 Loaded attestation provider's verification keys from %s.\n", issAddr+"/certs")

	// Parse token.
	token, err := jwt.ParseSigned(string(tokenBytes))
	if err != nil {
		panic(err)
	}

	// Verify token and get claims.
	var publicClaims jwt.Claims
	var privateClaims privateClaims
	if err := token.Claims(&keySet, &publicClaims, &privateClaims); err != nil {
		panic(err)
	}
	fmt.Println("✅ Token has a valid signature.")

	// Verify token claims.
	if err := verifyTokenClaims(publicClaims, privateClaims, *signer); err != nil {
		panic(err)
	}

	// Get certificate.
	certBytes, err := base64.RawURLEncoding.DecodeString(privateClaims.EnclaveHeldData)
	if err != nil {
		panic(err)
	}
	fmt.Println("🆗 Server certificate extracted from token.")

	// Create a TLS config that uses the server certificate as root
	// CA so that future connections to the server can be verified.
	cert, _ := x509.ParseCertificate(certBytes)
	tlsConfig = &tls.Config{RootCAs: x509.NewCertPool(), ServerName: "localhost"}
	tlsConfig.RootCAs.AddCert(cert)

	httpGet(tlsConfig, serverUrl+"/secret?s=thisIsTheSecret")
	fmt.Println("🔒 Sent secret over attested TLS channel.")
}

// verfiyTokenClaims verifies a bunch of important token claims.
// For productive envornments, please refere to the MS Azure Attestation Documentation
// and JWT as well as OpenID Connect specification to determine which parameters need
// to be verified.
func verifyTokenClaims(publicClaims jwt.Claims, privateClaims privateClaims, signer string) error {
	if err := publicClaims.Validate(jwt.Expected{Issuer: issAddr, Time: time.Now()}); err != nil {
		return fmt.Errorf("issuer of token could not be verified: %v", err)
	}
	fmt.Printf("✅ Issuer of token is %s.\n", publicClaims.Issuer)

	if privateClaims.SignerID != signer {
		return errors.New("token does not contain the right signer id")
	}
	fmt.Println("✅ SignerID of the toke equals the SignerID you passed to the client.")

	if privateClaims.ProductID != 1234 {
		return errors.New("token contains invalid product id")
	}
	fmt.Println("✅ Product ID verified.")

	if privateClaims.SecurityVersion < 2 {
		return errors.New("token contains invalid security version number")
	}
	fmt.Println("✅ Security Version Number verified.")

	return nil
}

// mustParseKeySet converts the raw keySetBytes to a JSONWebKeySet.
// This is needed only because the Shared Attestation Provider does not provide valid JWKs.
func mustParseKeySet(keySetBytes []byte) jose.JSONWebKeySet {
	var rawKeySet struct {
		Keys []struct {
			X5c []string
			Kid string
		}
	}
	if err := json.Unmarshal(keySetBytes, &rawKeySet); err != nil {
		panic(err)
	}

	var keySet jose.JSONWebKeySet
	for _, key := range rawKeySet.Keys {
		rawCert, _ := base64.StdEncoding.DecodeString(key.X5c[0])
		cert, err := x509.ParseCertificate(rawCert)
		if err != nil {
			panic(err)
		}
		keySet.Keys = append(keySet.Keys, jose.JSONWebKey{KeyID: key.Kid, Key: cert.PublicKey})
	}

	return keySet
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
