// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package attestation provides attestation data structures.
package attestation

import "errors"

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
