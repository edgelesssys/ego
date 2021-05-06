// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package attestation

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

func TestCreateAzureAttestationToken(t *testing.T) {
	require := require.New(t)
	//
	// Mock attestation provider.
	//
	createToken := func(w http.ResponseWriter, r *http.Request) {
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
		if string(report[:6]) != "111777" {
			http.Error(w, "invalid report", http.StatusBadRequest)
			return
		}
		if req.RuntimeData.Data == "" || req.RuntimeData.DataType == "" {
			http.Error(w, "missing runtime data", http.StatusBadRequest)
			return
		}
		token := attestationResponse{Token: "test"}
		response, err := json.Marshal(token)
		if err != nil {
			http.Error(w, "could not create reponse", http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
	attestationProvider := httptest.NewUnstartedServer(http.HandlerFunc(createToken))
	attestationProvider.Start()
	defer attestationProvider.Close()
	//
	// Test.
	//
	report := []byte("111777randombytes")
	data := []byte("test")
	baseurl := attestationProvider.URL
	resp, err := CreateAzureAttestationToken(report, data, baseurl)
	require.NoError(err)
	require.Equal(resp, "test")
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
	attestationProvider := httptest.NewUnstartedServer(http.HandlerFunc(serveKeys))
	attestationProvider.Start()
	defer attestationProvider.Close()
	//
	// Test cases.
	//
	tests := []struct {
		Name         string
		PublicClaims jwt.Claims
		Key          *rsa.PrivateKey
		ExpectErr    bool
	}{
		{
			"basic",
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    attestationProvider.URL,
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
			key, false,
		},
		{
			"wrong issuer",
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "wrong issuer",
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
			key, true,
		},
		{
			"expired",
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    attestationProvider.URL,
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
			key, true,
		},
		{
			"before nbf",
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Issuer:    attestationProvider.URL,
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			},
			key, true,
		},
		{
			"issued in future",
			jwt.Claims{
				Expiry:    jwt.NewNumericDate(time.Now().Add(4 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
				Issuer:    attestationProvider.URL,
				ID:        "222000",
				NotBefore: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			},
			key, true,
		},
		{
			"wrong signature",
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
	for _, test := range tests {
		t.Logf("Subtest: %v", test.Name)
		//
		// Create token.
		//
		publicClaims := test.PublicClaims
		privateClaims := privateClaims{
			Data:            base64.RawURLEncoding.EncodeToString(data),
			Debug:           true,
			UniqueID:        string(uniqueID),
			SignerID:        string(signerID),
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
		report, err := VerifyAzureAttestationToken(rawToken, attestationProvider.URL)
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

func TestSharedProviderKeyParsing(t *testing.T) {
	require := require.New(t)
	url := "https://shareduks.uks.attest.azure.net"
	tlsConfig := &tls.Config{}
	jwkSetBytes, err := httpGet(tlsConfig, url+"/certs")
	require.NoError(err)
	_, err = parseKeySet(jwkSetBytes)
	require.NoError(err)
}
