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

	"github.com/edgelesssys/ego/attestation"
	internal "github.com/edgelesssys/ego/internal/attestation"
)

// GetSelfReport returns a report of this enclave.
// The report can't be used for attestation, but to get values like the SignerID of this enclave.
func GetSelfReport() (attestation.Report, error) {
	// get empty local report to use it as target info
	report, err := GetLocalReport(nil, nil)
	if err != nil {
		return attestation.Report{}, err
	}
	// get report for target
	report, err = GetLocalReport(nil, report)
	if err != nil {
		return attestation.Report{}, err
	}
	// targeted report can be verified
	return VerifyLocalReport(report)
}

// CreateAttestationCertificate creates an X.509 certificate with an embedded report from the underlying enclave.
func CreateAttestationCertificate(template, parent *x509.Certificate, pub, priv interface{}) ([]byte, error) {
	return internal.CreateAttestationCertificate(GetRemoteReport, template, parent, pub, priv)
}

// CreateAttestationServerTLSConfig creates a tls.Config object with a self-signed certificate and an embedded report.
func CreateAttestationServerTLSConfig() (*tls.Config, error) {
	return internal.CreateAttestationServerTLSConfig(GetRemoteReport)
}

// CreateAzureAttestationToken creates a Microsoft Azure Attestation token by creating a
// remote report and sending it to an Attestation Provider, who is reachable under url.
// A JSON Web Token in compact serialization is returned.
func CreateAzureAttestationToken(data []byte, url string) (string, error) {
	hash := sha256.Sum256(data)
	report, err := GetRemoteReport(hash[:])
	if err != nil {
		return "", err
	}
	return internal.CreateAzureAttestationToken(report, data, url)
}
