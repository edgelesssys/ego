// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package eclient

import (
	"crypto/tls"

	"github.com/edgelesssys/ego/attestation"
	internal "github.com/edgelesssys/ego/internal/attestation"
)

// VerifyRemoteReport verifies the integrity of the remote report and its signature.
//
// This function verifies that the report signature is valid. It
// verifies that the signing authority is rooted to a trusted authority
// such as the enclave platform manufacturer.
//
// The caller must verify the returned report's content.
func VerifyRemoteReport(reportBytes []byte) (attestation.Report, error) {
	report, err := verifyRemoteReport(reportBytes)
	return toAttestationReport(report), err
}

// CreateAttestationClientTLSConfig creates a tls.Config object that verifies a certificate with embedded report.
//
// verifyReport is called after the certificate has been verified against the report data. The caller must verify either the UniqueID or the tuple (SignerID, ProductID, SecurityVersion) in the callback.
func CreateAttestationClientTLSConfig(verifyReport func(attestation.Report) error) *tls.Config {
	return internal.CreateAttestationClientTLSConfig(verifyRemoteReport, func(report internal.Report) error {
		return verifyReport(toAttestationReport(report))
	})
}

func toAttestationReport(report internal.Report) attestation.Report {
	return attestation.Report{
		Data:            report.Data,
		SecurityVersion: report.SecurityVersion,
		Debug:           report.Debug,
		UniqueID:        report.UniqueID,
		SignerID:        report.SignerID,
		ProductID:       report.ProductID}
}
