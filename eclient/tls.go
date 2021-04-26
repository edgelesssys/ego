// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package eclient

import (
	"crypto/tls"

	internal "github.com/edgelesssys/ego/internal/attestation"
)

// CreateAttestationClientTLSConfig creates a tls.Config object that verifies a certificate with embedded report.
//
// verifyReport is called after the certificate has been verified against the report data. The caller must verify either the UniqueID or the tuple (SignerID, ProductID, SecurityVersion) in the callback.
func CreateAttestationClientTLSConfig(verifyReport func(internal.Report) error) *tls.Config {
	return internal.CreateAttestationClientTLSConfig(VerifyRemoteReport, verifyReport)
}
