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

// GetSealKeyID gets a unique ID derived from the CPU's root seal key.
// The ID also depends on the ProductID and Debug flag of the enclave.
func GetSealKeyID() ([]byte, error) {
	// see https://github.com/intel/linux-sgx/blob/sgx_2.19/common/inc/sgx_key.h
	keyRequest := make([]byte, 512)
	keyRequest[0] = 4 // key_name = SGX_KEYSELECT_SEAL
	// Leaving all other fields 0 means only ProductID (but not unique or signer id) of the enclave
	// is used for derivation. The key isn't secret because everyone can create an enclave with the
	// same ProductID that prints the key. So we can directly use it as an ID for the CPU.
	return GetSealKey(keyRequest)
}

// CreateAttestationCertificate creates an X.509 certificate with an embedded report from the underlying enclave.
func CreateAttestationCertificate(template, parent *x509.Certificate, pub, priv interface{}) ([]byte, error) {
	return internal.CreateAttestationCertificate(GetRemoteReport, template, parent, pub, priv)
}

// CreateAttestationServerTLSConfig creates a tls.Config object with a self-signed certificate and an embedded report.
func CreateAttestationServerTLSConfig() (*tls.Config, error) {
	return internal.CreateAttestationServerTLSConfig(GetRemoteReport)
}

// CreateAttestationClientTLSConfig creates a tls.Config object that verifies a certificate with embedded report.
//
// verifyReport is called after the certificate has been verified against the report data. The caller must verify either the UniqueID or the tuple (SignerID, ProductID, SecurityVersion, Debug) in the callback.
func CreateAttestationClientTLSConfig(verifyReport func(attestation.Report) error, opts ...AttestOption) *tls.Config {
	var appliedOpts internal.Options
	for _, o := range opts {
		o.apply(&appliedOpts)
	}

	return internal.CreateAttestationClientTLSConfig(
		func(reportBytes []byte) (internal.Report, error) {
			report, err := VerifyRemoteReport(reportBytes)
			return internal.Report(report), err
		},
		appliedOpts,
		func(rep internal.Report) error { return verifyReport(attestation.Report(rep)) },
	)
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

// AttestOption	configures an attestation function.
type AttestOption struct {
	apply func(*internal.Options)
}

// WithIgnoreTCBStatus ignores an invalid TCB level.
//
// Callers must verify the TCBStatus field in the report themselves.
func WithIgnoreTCBStatus() AttestOption {
	return AttestOption{func(o *internal.Options) { o.IgnoreErr = attestation.ErrTCBLevelInvalid }}
}
