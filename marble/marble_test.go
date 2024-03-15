// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package marble

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTLSConfig(t *testing.T) {
	defer resetEnv()
	assert := assert.New(t)
	require := require.New(t)

	// Get server TLS config
	setupTest(require)
	tlsConfig, err := GetTLSConfig(true)
	require.NoError(err)
	assert.NotNil(tlsConfig)

	// Check root certificate
	caCommonName, err := getCommonNameFromX509Pool(tlsConfig.RootCAs)
	require.NoError(err)
	assert.Equal("Test CA", caCommonName)

	// Check client CA certificate
	caCommonName, err = getCommonNameFromX509Pool(tlsConfig.ClientCAs)
	require.NoError(err)
	assert.Equal("Test CA", caCommonName)

	// Check leaf certificate
	certificates := tlsConfig.Certificates
	rawCertificate := certificates[0].Certificate[0]
	x509Cert, err := x509.ParseCertificate(rawCertificate)
	require.NoError(err)
	assert.Equal(big.NewInt(1337), x509Cert.SerialNumber)
	assert.Equal("Test Leaf", x509Cert.Subject.CommonName)

	// Check ClientAuth value
	assert.Equal(tls.RequireAndVerifyClientCert, tlsConfig.ClientAuth)

	// Check ClientAuth value when false is used as parameter
	tlsConfigNoClientAuth, err := GetTLSConfig(false)
	require.NoError(err)
	assert.Nil(tlsConfigNoClientAuth.ClientCAs)
	assert.NotEqual(tls.RequireAndVerifyClientCert, tlsConfigNoClientAuth.ClientAuth)
}

func TestGarbageEnviromentVars(t *testing.T) {
	defer resetEnv()
	assert := assert.New(t)

	// Set environment variables
	os.Setenv(MarbleEnvironmentRootCA, "this")
	os.Setenv(MarbleEnvironmentCertificateChain, "is")
	os.Setenv(MarbleEnvironmentPrivateKey, "some serious garbage")

	// This should fail
	tlsConfig, err := GetTLSConfig(true)
	assert.Error(err)
	assert.Nil(tlsConfig)
}

func TestMissingEnvironmentVars(t *testing.T) {
	assert := assert.New(t)
	tlsConfig, err := GetTLSConfig(false)

	assert.Error(err)
	assert.Nil(tlsConfig)
}

func setupTest(require *require.Assertions) {
	// Generate keys
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(err)
	privKey, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(err)

	// Create some demo CA certificate
	templateCa := x509.Certificate{
		SerialNumber: big.NewInt(42),
		IsCA:         true,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
	}

	// Create some demo leaf certificate
	templateLeaf := x509.Certificate{
		SerialNumber: big.NewInt(1337),
		IsCA:         false,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour * 24 * 365),
		Subject: pkix.Name{
			CommonName: "Test Leaf",
		},
	}

	// Create test CA cert
	certCaRaw, err := x509.CreateCertificate(rand.Reader, &templateCa, &templateCa, &key.PublicKey, key)
	require.NoError(err)

	certCa, err := x509.ParseCertificate(certCaRaw)
	require.NoError(err)

	// Create test leaf cert
	certLeafRaw, err := x509.CreateCertificate(rand.Reader, &templateLeaf, certCa, &key.PublicKey, key)
	require.NoError(err)

	certLeaf, err := x509.ParseCertificate(certLeafRaw)
	require.NoError(err)

	// Convert them to PEM
	caCertPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certCa.Raw})
	leafCertPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certLeaf.Raw})
	privKeyPem := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privKey})

	// Set environment variables
	os.Setenv(MarbleEnvironmentRootCA, string(caCertPem))
	os.Setenv(MarbleEnvironmentCertificateChain, string(leafCertPem))
	os.Setenv(MarbleEnvironmentPrivateKey, string(privKeyPem))
}

func resetEnv() {
	// Clean up used environment variables, otherwise they stay set!
	os.Unsetenv(MarbleEnvironmentRootCA)
	os.Unsetenv(MarbleEnvironmentCertificateChain)
	os.Unsetenv(MarbleEnvironmentPrivateKey)
}

// x509 cert pools don't allow to extract certificates inside them. How great is that? So we gotta extract the ASN.1 subject and work with it.
// This was taken (and slightly modified) from: https://github.com/golang/go/issues/26614#issuecomment-613640345
func getCommonNameFromX509Pool(pool *x509.CertPool) (string, error) {
	poolSubjects := pool.Subjects() //nolint:staticcheck

	var rdnSequence pkix.RDNSequence
	_, err := asn1.Unmarshal(poolSubjects[0], &rdnSequence)
	if err != nil {
		return "", err
	}

	var name pkix.Name
	name.FillFromRDNSequence(&rdnSequence)
	return name.CommonName, nil
}
