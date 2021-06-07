// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package marble provides commonly used functionalities for Marblerun Marbles.
package marble

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// MarbleEnvironmentCertificateChain contains the name of the environment variable holding a marble-specifc PEM encoded certificate
const MarbleEnvironmentCertificateChain = "MARBLE_PREDEFINED_MARBLE_CERTIFICATE_CHAIN"

// MarbleEnvironmentRootCA contains the name of the environment variable holding a PEM encoded root certificate
const MarbleEnvironmentRootCA = "MARBLE_PREDEFINED_ROOT_CA"

// MarbleEnvironmentPrivateKey contains the name of the environment variable holding a PEM encoded private key belonging to the marble-specific certificate
const MarbleEnvironmentPrivateKey = "MARBLE_PREDEFINED_PRIVATE_KEY"

// GetTLSConfig provides a preconfigured TLS config for marbles, using the Marblerun Coordinator as trust anchor
func GetTLSConfig(verifyClientCerts bool) (*tls.Config, error) {
	tlsCert, roots, err := generateFromEnv()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs:      roots,
		Certificates: []tls.Certificate{tlsCert},
	}

	if verifyClientCerts {
		tlsConfig.ClientCAs = roots
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}

func getByteEnv(name string) ([]byte, error) {
	value := os.Getenv(name)
	if len(value) == 0 {
		return nil, fmt.Errorf("environment variable not set: %s", name)
	}
	return []byte(value), nil
}

func generateFromEnv() (tls.Certificate, *x509.CertPool, error) {
	certChain, err := getByteEnv(MarbleEnvironmentCertificateChain)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	rootCA, err := getByteEnv(MarbleEnvironmentRootCA)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	leafPrivk, err := getByteEnv(MarbleEnvironmentPrivateKey)
	if err != nil {
		return tls.Certificate{}, nil, err
	}

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(rootCA) {
		return tls.Certificate{}, nil, fmt.Errorf("cannot append rootCA to CertPool")
	}

	tlsCert, err := tls.X509KeyPair(certChain, leafPrivk)
	if err != nil {
		return tls.Certificate{}, nil, fmt.Errorf("cannot create TLS cert: %v", err)
	}

	return tlsCert, roots, nil
}
