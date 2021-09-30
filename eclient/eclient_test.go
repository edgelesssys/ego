// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package eclient

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/edgelesssys/ego/attestation"
)

func ExampleCreateAttestationClientTLSConfig() {
	// the uniqueID is derived from the binary of the enclaved program
	// and can be obtained using `ego uniqueid`
	var uniqueID []byte

	verifyReport := func(report attestation.Report) error {
		if !bytes.Equal(report.UniqueID, uniqueID) {
			return errors.New("invalid UniqueID")
		}
		return nil
	}

	tlsConfig := CreateAttestationClientTLSConfig(verifyReport)
	client := http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}

	// example.com must use a TLS certificate with an embedded report
	// EGo's enclave package provides functionality for such server
	_, _ = client.Get("https://example.com")
}
