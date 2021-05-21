// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package enclave

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"

	"github.com/edgelesssys/ego/internal/attestation"
)

// CreateAttestationCertificate creates an X.509 certificate with an embedded report from the underlying enclave.
func CreateAttestationCertificate(template, parent *x509.Certificate, pub, priv interface{}) ([]byte, error) {
	return attestation.CreateAttestationCertificate(GetRemoteReport, template, parent, pub, priv)
}

// CreateAttestationServerTLSConfig creates a tls.Config object with a self-signed certificate and an embedded report.
func CreateAttestationServerTLSConfig() (*tls.Config, error) {
	return attestation.CreateAttestationServerTLSConfig(GetRemoteReport)
}

// CreateAzureAttestationToken creates a Microsoft Azure Attestation Token by creating an
// remote report and sending the report to an Attestation Provider, who is reachable
// under baseurl. The Attestation Provider will verify the remote Report.
// A JSON Web Token in compact serialization is returned.
func CreateAzureAttestationToken(data []byte, url string) (string, error) {
	hash := sha256.Sum256(data)
	report, err := GetRemoteReport(hash[:])
	if err != nil {
		return "", err
	}
	return attestation.CreateAzureAttestationToken(report, data, url)
}
