// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package eclient provides functionality for Go programs that interact with enclave programs.
package eclient

// #cgo LDFLAGS: -loehostverify -lcrypto -ldl
// #include <openenclave/host_verify.h>
import "C"

import (
	"errors"
	"unsafe"

	"github.com/edgelesssys/ego/attestation"
	internal "github.com/edgelesssys/ego/internal/attestation"
)

// VerifyRemoteReport verifies the integrity of the remote report and its signature.
//
// This function verifies that the report signature is valid. It
// verifies that the signing authority is rooted to a trusted authority
// such as the enclave platform manufacturer.
//
// Returns the parsed report if the signature is valid.
// Returns an error if the signature is invalid.
func VerifyRemoteReport(reportBytes []byte) (internal.Report, error) {
	if len(reportBytes) <= 0 {
		return internal.Report{}, attestation.ErrEmptyReport
	}

	var report C.oe_report_t

	res := C.oe_verify_remote_report(
		(*C.uint8_t)(&reportBytes[0]), C.size_t(len(reportBytes)),
		nil, 0,
		&report)

	if res != C.OE_OK {
		return internal.Report{}, oeError(res)
	}

	return internal.Report{
		Data:            C.GoBytes(unsafe.Pointer(report.report_data), C.int(report.report_data_size)),
		SecurityVersion: uint(report.identity.security_version),
		Debug:           (report.identity.attributes & C.OE_REPORT_ATTRIBUTES_DEBUG) != 0,
		UniqueID:        C.GoBytes(unsafe.Pointer(&report.identity.unique_id[0]), C.OE_UNIQUE_ID_SIZE),
		SignerID:        C.GoBytes(unsafe.Pointer(&report.identity.signer_id[0]), C.OE_SIGNER_ID_SIZE),
		ProductID:       C.GoBytes(unsafe.Pointer(&report.identity.product_id[0]), C.OE_PRODUCT_ID_SIZE),
	}, nil
}

func oeError(res C.oe_result_t) error {
	return errors.New(C.GoString(C.oe_result_str(res)))
}
