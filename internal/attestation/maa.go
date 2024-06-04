// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package attestation

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// CreateAzureAttestationToken creates a Microsoft Azure Attestation Token by sending a report
// to an Attestation Provider, who is reachable under baseurl. A JSON Web Token in compact
// serialization is returned.
func CreateAzureAttestationToken(report, data []byte, baseurl string) (string, error) {
	// Create attestation request struct.
	rtd := rtdata{Data: base64.RawURLEncoding.EncodeToString(data), DataType: "Binary"}
	attReq := attestOERequest{Report: base64.RawURLEncoding.EncodeToString(report), RuntimeData: rtd}

	// Parse url and add path.
	uri, err := url.Parse(baseurl)
	if err != nil {
		return "", err
	}
	path, err := url.Parse("/attest/OpenEnclave?api-version=2020-10-01")
	if err != nil {
		return "", err
	}
	uri = uri.ResolveReference(path)

	// Marshal request struct to JSON.
	jsonReq, err := json.Marshal(attReq)
	if err != nil {
		return "", err
	}

	// Create HTTP client skiping TLS certificate verification, since
	// the enclave does not have a set of Root CAs. There is no need
	// for a trusted connection.
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	resp, err := client.Post(uri.String(), "application/json", bytes.NewReader(jsonReq))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check response and return the token.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("attestation request failed, attestation provider returned status code %v", resp.StatusCode)
	}
	body := new(attestationResponse)
	if err := json.NewDecoder(resp.Body).Decode(body); err != nil {
		return "", err
	}
	return body.Token, nil
}

// VerifyAzureAttestationToken takes a Microsoft Azure Attestation Token in JSON Web Token compact
// serialization format and verifies the tokens public claims and signature. The Attestation providers
// keys are loaded from baseURL over an TLS connection. The validation is based on the trust in this TLS channel.
// Note, that the token's issuer (iss) has to equal the baseURL's string representation.
//
// Attention: the calling function needs to ensure the scheme of baseURL is HTTPS,
// e.g. by calling the ParseHTTPS function of this package.
func VerifyAzureAttestationToken(rawToken string, baseURL *url.URL) (Report, error) {
	// Parse baseURL and add path.
	path, err := url.Parse("/certs")
	if err != nil {
		return Report{}, err
	}
	uri := baseURL.ResolveReference(path)

	// Get JSON Web Key set.
	jwkSetBytes, err := httpGet(uri.String())
	if err != nil {
		return Report{}, err
	}
	keySet, err := parseKeySet(jwkSetBytes)
	if err != nil {
		return Report{}, err
	}

	// Parse token.
	token, err := jwt.ParseSigned(rawToken, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return Report{}, err
	}

	// Verify token and get claims.
	var publicClaims jwt.Claims
	var privateClaims privateClaims
	if err := token.Claims(&keySet, &publicClaims, &privateClaims); err != nil {
		return Report{}, err
	}

	// Verify public claims.
	if err := publicClaims.Validate(jwt.Expected{Issuer: baseURL.String(), Time: time.Now()}); err != nil {
		return Report{}, err
	}

	// Create a report form the private claims.
	data, err := base64.RawURLEncoding.DecodeString(privateClaims.Data)
	if err != nil {
		return Report{}, err
	}
	uniqueID, err := hex.DecodeString(privateClaims.UniqueID)
	if err != nil {
		return Report{}, err
	}
	signerID, err := hex.DecodeString(privateClaims.SignerID)
	if err != nil {
		return Report{}, err
	}
	productID := make([]byte, 16)
	binary.LittleEndian.PutUint16(productID, uint16(privateClaims.ProductID))
	return Report{
		Data:            data,
		SecurityVersion: privateClaims.SecurityVersion,
		Debug:           privateClaims.Debug,
		UniqueID:        uniqueID,
		SignerID:        signerID,
		ProductID:       productID,
	}, nil
}

// ParseHTTPS parses an URL and ensures its scheme is HTTPS.
func ParseHTTPS(URL string) (*url.URL, error) {
	uri, err := url.Parse(URL)
	if err != nil {
		return nil, errors.New("could not parse URL")
	}
	if uri.Scheme != "https" {
		return nil, errors.New("the provided baseURL does not use HTTPS")
	}
	return uri, nil
}

// atttestOERequest is an Microsoft Azure Attestation AttestOpenEnclaveRequest.
// See https://docs.microsoft.com/en-us/rest/api/attestation/attestation/attestopenenclave
// for REST API documentation of Azure Attestation Provider.
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

// privateClaims are some of the private claims of an Azure Attestation token.
type privateClaims struct {
	Data            string `json:"x-ms-sgx-ehd"`
	SecurityVersion uint   `json:"x-ms-sgx-svn"`
	Debug           bool   `json:"x-ms-sgx-is-debuggable"`
	UniqueID        string `json:"x-ms-sgx-mrenclave"`
	SignerID        string `json:"x-ms-sgx-mrsigner"`
	ProductID       uint   `json:"x-ms-sgx-product-id"`
}

func parseKeySet(keySetBytes []byte) (jose.JSONWebKeySet, error) {
	var rawKeySet struct {
		Keys []struct {
			X5c []string
			Kid string
		}
	}
	if err := json.Unmarshal(keySetBytes, &rawKeySet); err != nil {
		return jose.JSONWebKeySet{}, err
	}

	var keySet jose.JSONWebKeySet
	for _, key := range rawKeySet.Keys {
		rawCert, _ := base64.StdEncoding.DecodeString(key.X5c[0])
		cert, err := x509.ParseCertificate(rawCert)
		if err != nil {
			return jose.JSONWebKeySet{}, err
		}
		keySet.Keys = append(keySet.Keys, jose.JSONWebKey{KeyID: key.Kid, Key: cert.PublicKey})
	}

	return keySet, nil
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http response has status %v", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
