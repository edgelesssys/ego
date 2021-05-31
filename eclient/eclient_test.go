// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package eclient

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net/http"

	"github.com/edgelesssys/ego/attestation"
)

func ExampleCreateAttestationClientTLSConfig() {
	// the signerID is derived from the binary of the enclaved program
	// and can be obtained using `ego signerid`
	var signerID []byte

	// verifyReport verifies the tuple (SignerID, ProductID, SecurityVersion)
	verifyReport := func(report attestation.Report) error {
		if report.SecurityVersion < 2 {
			return errors.New("invalid security version")
		}
		if binary.LittleEndian.Uint16(report.ProductID) != 1234 {
			return errors.New("invalid product")
		}
		if !bytes.Equal(report.SignerID, signerID) {
			return errors.New("invalid signer")
		}
		return nil
	}

	tlsConfig := CreateAttestationClientTLSConfig(verifyReport)
	client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}

	// example.com must use a TLS certificate with an embedded report
	// EGo's enclave package provides functionality for such server
	_, _ = client.Get("https://example.com")
}
