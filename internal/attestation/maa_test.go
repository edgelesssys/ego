// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package attestation

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

func TestCreateAzureAttestationToken(t *testing.T) {
	require := require.New(t)
	//
	// Test cases.
	//
	tests := map[string]struct {
		report        []byte
		data          []byte
		expectedToken string
		expectedError error
	}{
		"basic": {
			[]byte("reportBytes"),
			[]byte("dataBytes"),
			"token",
			nil,
		},
	}
	for testName, test := range tests {
		t.Logf("Subtest: %v", testName)
		//
		// Mock attestation provider.
		//
		createToken := func(w http.ResponseWriter, r *http.Request) {
			if r.RequestURI != "/attest/OpenEnclave?api-version=2020-10-01" {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}

			decoder := json.NewDecoder(r.Body)
			var req attestOERequest
			if err := decoder.Decode(&req); err != nil {
				http.Error(w, "could not decode json", http.StatusBadRequest)
				return
			}
			report, err := base64.RawURLEncoding.DecodeString(req.Report)
			if err != nil {
				http.Error(w, "could not decode report", http.StatusBadRequest)
				return
			}

			if !bytes.Equal(report, test.report) {
				http.Error(w, "invalid report", http.StatusBadRequest)
				return
			}
			if req.RuntimeData.Data != base64.RawURLEncoding.EncodeToString(test.data) {
				http.Error(w, "invalid runtime data", http.StatusBadRequest)
				return
			}
			if req.RuntimeData.DataType != "Binary" {
				http.Error(w, "runtime data of invalid type", http.StatusBadRequest)
				return
			}

			token := attestationResponse{Token: test.expectedToken}
			if err := json.NewEncoder(w).Encode(token); err != nil {
				http.Error(w, "could not create response", http.StatusInternalServerError)
			}
		}
		//
		// Run test.
		//
		func() {
			attestationProvider := httptest.NewServer(http.HandlerFunc(createToken))
			defer attestationProvider.Close()

			resp, err := CreateAzureAttestationToken(test.report, test.data, attestationProvider.URL)
			if test.expectedError != nil {
				require.Error(err)
			} else {
				require.NoError(err)
				require.Equal(resp, test.expectedToken)
			}
		}()
	}
}

func TestVerifyAzureAttestationToken(t *testing.T) {
	require := require.New(t)
	//
	// Test values.
	//
	data := []byte("test")
	uniqueID := []byte("222111")
	signerID := []byte("222333")
	//
	// Create cert.
	//
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(err)
	cert := &x509.Certificate{
		SerialNumber: &big.Int{},
		Subject:      pkix.Name{CommonName: "attestation provider"},
		DNSNames:     []string{"localhost"},
		NotAfter:     time.Now().Add(time.Hour),
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	require.NoError(err)
	// evilKey
	evilKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(err)
	//
	// Create JWK.
	//
	type rawKey struct {
		X5c []string `json:"x5c"`
		Kty string   `json:"kty"`
		Kid string   `json:"kid"`
	}
	jwk := rawKey{
		Kty: "RSA",
		Kid: "aaa",
		X5c: []string{base64.StdEncoding.EncodeToString(certBytes)},
	}
	rawKeySet := struct {
		Keys []rawKey `json:"keys"`
	}{
		Keys: []rawKey{jwk},
	}
	//
	// Mock key server.
	//
	serveKeys := func(w http.ResponseWriter, r *http.Request) {
		response, err := json.Marshal(rawKeySet)
		if err != nil {
			http.Error(w, "could not marshal json keys", http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
	attestationProvider := httptest.NewServer(http.HandlerFunc(serveKeys))
	defer attestationProvider.Close()
	//
	// Test cases.
	//
	tests := map[string]struct {
		PublicClaims jwt.Claims
		Key          *rsa.PrivateKey
		ExpectErr    bool
	}{
		"basic": {
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    attestationProvider.URL,
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
			key, false,
		},
		"wrong issuer": {
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "wrong issuer",
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
			key, true,
		},
		"expired": {
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    attestationProvider.URL,
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
			key, true,
		},
		"before nbf": {
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    attestationProvider.URL,
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
			key, true,
		},
		"issued in future": {
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(4 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
				Issuer:    attestationProvider.URL,
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			},
			key, true,
		},
		"wrong signature": {
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(4 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
				Issuer:    attestationProvider.URL,
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			},
			evilKey, true,
		},
	}
	for testName, test := range tests {
		t.Logf("Subtest: %v", testName)
		//
		// Create token.
		//
		publicClaims := test.PublicClaims
		privateClaims := privateClaims{
			Data:            base64.RawURLEncoding.EncodeToString(data),
			Debug:           true,
			UniqueID:        hex.EncodeToString(uniqueID),
			SignerID:        hex.EncodeToString(signerID),
			ProductID:       123,
			SecurityVersion: 321,
		}
		sig, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: test.Key}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "aaa"))
		require.NoError(err)
		rawToken, err := jwt.Signed(sig).Claims(publicClaims).Claims(privateClaims).CompactSerialize()
		require.NoError(err)
		//
		// Verify token and check report.
		//
		uri, err := url.Parse(attestationProvider.URL)
		require.NoError(err)
		report, err := VerifyAzureAttestationToken(rawToken, uri)
		if test.ExpectErr {
			require.Error(err)
			continue
		} else {
			require.NoError(err)
		}
		require.Equal(data, report.Data)
		require.Equal(privateClaims.SecurityVersion, report.SecurityVersion)
		require.Equal(privateClaims.Debug, report.Debug)
		require.Equal(uniqueID, report.UniqueID)
		require.Equal(signerID, report.SignerID)
		productID := make([]byte, 16)
		binary.LittleEndian.PutUint16(productID, uint16(privateClaims.ProductID))
		require.Equal(productID, report.ProductID)
	}
}

func DisabledTestSharedProviderKeyParsing(t *testing.T) {
	require := require.New(t)
	url := "https://shareduks.uks.attest.azure.net"
	jwkSetBytes, err := httpGet(url + "/certs")
	require.NoError(err)
	_, err = parseKeySet(jwkSetBytes)
	require.NoError(err)
}

func TestParseHTTPS(t *testing.T) {
	require := require.New(t)

	invalidURLs := []string{
		"http://example.com",
		"ftp://example.com",
	}
	for _, url := range invalidURLs {
		_, err := ParseHTTPS(url)
		require.Error(err)
	}

	validURLs := []string{
		"https://example.com",
		"https://shareduks.uks.attest.azure.net",
	}
	for _, url := range validURLs {
		_, err := ParseHTTPS(url)
		require.NoError(err)
	}
}
