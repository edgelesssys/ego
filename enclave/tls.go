// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package enclave

import (
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
