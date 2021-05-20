// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package attestation provides attestation data structures.
package attestation

import (
	"errors"

	"github.com/edgelesssys/ego/internal/attestation"
)

// Report is a parsed enclave report.
type Report struct {
	Data            []byte // The report data that has been included in the report.
	SecurityVersion uint   // Security version of the enclave. For SGX enclaves, this is the ISVSVN value.
	Debug           bool   // If true, the report is for a debug enclave.
	UniqueID        []byte // The unique ID for the enclave. For SGX enclaves, this is the MRENCLAVE value.
	SignerID        []byte // The signer ID for the enclave. For SGX enclaves, this is the MRSIGNER value.
	ProductID       []byte // The Product ID for the enclave. For SGX enclaves, this is the ISVPRODID value.
}

// ErrEmptyReport is returned by VerifyRemoteReport if reportBytes is empty.
var ErrEmptyReport = errors.New("empty report")

// VerifyAzureAttestationToken takes a Microsoft Azure Attestation Token in JSON Web Token compact
// serialization format and verifies the tokens public claims and signature. The Attestation providers
// keys are loaded from providerURL/certs over TLS and need to be in JSON Web Key format. The validation
// is based on the trust in this TLS channel. Note, that the token's issuer (iss) has to equal the providerURL.
//
// Since an enclave hasn't got a set of root CA certificates, there is no default way to establish a
// trusted TLS connection to the Attestation Provider for getting the certificate the token was signed with.
// This function is therefore usable by non-enclaved clients only.
func VerifyAzureAttestationToken(token string, providerURL string) (Report, error) {
	// Ensure providerURL uses HTTPS.
	uri, err := attestation.ParseHTTPS(providerURL)
	if err != nil {
		return Report{}, err
	}
	report, err := attestation.VerifyAzureAttestationToken(token, uri)
	if err != nil {
		return Report{}, err
	}
	return Report{
		Data:            report.Data,
		SecurityVersion: report.SecurityVersion,
		Debug:           report.Debug,
		UniqueID:        report.UniqueID,
		SignerID:        report.SignerID,
		ProductID:       report.ProductID}, nil
}
