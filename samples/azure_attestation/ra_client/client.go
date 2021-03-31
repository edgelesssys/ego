package main

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// issAddr is the URL of the attestation provider
const issAddr = "https://shareduks.uks.attest.azure.net"

func main() {
	signer := flag.String("s", "", "signer ID")
	serverAddr := flag.String("a", "localhost:8080", "server address")
	flag.Parse()

	// Ensure signer ID is passed.
	if len(*signer) == 0 {
		flag.Usage()
		return
	}

	// Get server certificate and its token. Skip TLS certificate verification because
	// the certificate is self-signed and we will verify it using the token instead.
	serverUrl := "https://" + *serverAddr
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	tokenBytes := httpGet(tlsConfig, serverUrl+"/token")
	fmt.Printf("🆗 Loaded server attestation token from %s.\n", serverUrl+"/token")

	// Get cerificate for token verification from attestation provider.
	tlsConfig = &tls.Config{}
	jwkSetBytes := httpGet(tlsConfig, issAddr+"/certs")
	var keySet JSONWebKeySet
	if err := json.Unmarshal(jwkSetBytes, &keySet); err != nil {
		panic(err)
	}
	fmt.Printf("🆗 Loaded attestation provider's verification keys from %s.\n", issAddr+"/certs")

	// Parse token.
	parts := strings.Split(string(tokenBytes), ".")
	for i, part := range parts {
		parts[i] = string(base64URLDecode(part))
	}
	var header JOSEHeader
	if err := json.Unmarshal([]byte(parts[0]), &header); err != nil {
		panic(err)
	}

	// Find signing key.
	key, err := keySet.findKid(header.Kid)
	if err != nil {
		panic(err)
	}
	fmt.Println("🆗 Found the attestation provider's key which was used to sign the token.")

	// Convert key to PEM, since the provided key of this instance are not valid JSON Web Keys
	pubk := jwkToRSAPub(key)

	// Parse token.
	getKey := func(token *jwt.Token) (interface{}, error) { return pubk, nil }
	token, err := jwt.Parse(string(tokenBytes), getKey)
	if err != nil {
		panic(err)
	}
	claims := token.Claims.(jwt.MapClaims)

	// Verfify token.
	mustVerifyTokenClaims(claims, *signer)

	// Verify certificate.
	certBytes := base64URLDecode(claims["x-ms-sgx-ehd"].(string))
	fmt.Println("🆗 Server certificate extracted from token.")

	// Create a TLS config that uses the server certificate as root
	// CA so that future connections to the server can be verified.
	cert, _ := x509.ParseCertificate(certBytes)
	tlsConfig = &tls.Config{RootCAs: x509.NewCertPool(), ServerName: "localhost"}
	tlsConfig.RootCAs.AddCert(cert)

	httpGet(tlsConfig, serverUrl+"/secret?s=thisIsTheSecret")
	fmt.Println("🔒 Sent secret over attested TLS channel.")
}

// mustVerfiyTokenClaims verifies a bunch of important token claims.
// For productive envornments, please refere to the MS Azure Attestation Documentation
// and JWT as well as OpenID Connect specification to determine which parameters need
// to be verified.
func mustVerifyTokenClaims(claims jwt.MapClaims, signer string) {
	if err := claims.Valid(); err != nil {
		panic(err)
	}
	fmt.Println("✅ Token is valid.")

	if !claims.VerifyIssuer(issAddr, true) {
		panic(errors.New("issuer of token could not be verified"))
	}
	fmt.Printf("✅ Issuer of token is %s.\n", claims["iss"])

	if claims["x-ms-sgx-mrsigner"] != signer {
		panic(errors.New("token does not contain the right signer id"))
	}
	fmt.Println("✅ SignerID of the toke equals the SignerID you passed to the client.")

	if claims["x-ms-sgx-product-id"] != float64(1234) {
		panic(errors.New("token contains invalid product id"))
	}
	fmt.Println("✅ Product ID verified.")

	svn, ok := claims["x-ms-sgx-svn"].(float64)
	if !ok {
		panic(errors.New("could not get security version number"))
	}
	if svn < float64(2) {
		panic(errors.New("token contains invalid security version number"))
	}
	fmt.Println("✅ Security Version Number verified.")
}

// jwkToRSAPub extracts a rsa.PublicKey form a JWK's x5u parameter.
// This is needed only because the Shared Attestation Provider does not provide valid JWKs.
func jwkToRSAPub(key *JSONWebKey) *rsa.PublicKey {
	pemData := key.getPEM()
	var decodedCertBytes []byte
	hex.Decode(decodedCertBytes, []byte(key.Certificates[0]))
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		panic(errors.New("failed to decode PEM block containing public key"))
	}
	certAP, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	return certAP.PublicKey.(*rsa.PublicKey)
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

// JSONWebKeySet is a set of JSONWebKey.
type JSONWebKeySet struct {
	Keys []JSONWebKey `json:"keys"`
}

// findKid searches a JSONWebKeySet for a key with a given key id.
func (set JSONWebKeySet) findKid(id string) (*JSONWebKey, error) {
	for _, key := range set.Keys {
		if key.KeyID == id {
			return &key, nil
		}
	}
	return nil, errors.New("could not find kid")
}

// JSONWebKey represents a JSON Web Key (JWK).
type JSONWebKey struct {
	Certificates []string `json:"x5c"`
	KeyID        string   `json:"kid"`
	KeyType      string   `json:"kty"`
}

// getPEM convertrs a JSONWebKey from DER format to PEM.
func (key JSONWebKey) getPEM() string {
	pemStart := "-----BEGIN CERTIFICATE-----"
	x5c := key.Certificates[0]
	pemEnd := "-----END CERTIFICATE-----"
	return fmt.Sprintf("%s\n%s\n%s", pemStart, x5c, pemEnd)
}

// JOSEHeader represents the JOSE header of a JWT
type JOSEHeader struct {
	Alg string `json:"alg"`
	Jku string `json:"jku"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

func base64URLDecode(s string) []byte {
	s = strings.Replace(s, "-", "+", -1)
	s = strings.Replace(s, "_", "/", -1)
	padlen := (4 - len(s)%4) % 4
	s += strings.Repeat("=", padlen)
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}
