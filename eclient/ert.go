// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build !ego_mock_eclient

package eclient

// #cgo LDFLAGS: -loehostverify -lcrypto -ldl
// #include <openenclave/attestation/verifier.h>
import "C"

import (
	"errors"
	"unsafe"

	"github.com/edgelesssys/ego/attestation"
	internal "github.com/edgelesssys/ego/internal/attestation"
)

func verifyRemoteReport(reportBytes []byte) (internal.Report, error) {
	if len(reportBytes) <= 0 {
		return internal.Report{}, attestation.ErrEmptyReport
	}

	res := C.oe_verifier_initialize()
	if res != C.OE_OK {
		return internal.Report{}, oeError(res)
	}

	var claims *C.oe_claim_t
	var claimsLength C.size_t

	res = C.oe_verify_evidence(
		nil,
		(*C.uint8_t)(&reportBytes[0]), C.size_t(len(reportBytes)),
		nil, 0,
		nil, 0,
		&claims, &claimsLength,
	)

	var verifyErr error
	if res == C.OE_TCB_LEVEL_INVALID {
		verifyErr = attestation.ErrTCBLevelInvalid
	} else if res != C.OE_OK {
		return internal.Report{}, oeError(res)
	}

	defer C.oe_free_claims(claims, claimsLength)

	report, err := internal.ParseClaims(uintptr(unsafe.Pointer(claims)), uintptr(claimsLength))
	if err != nil {
		return internal.Report{}, err
	}
	return report, verifyErr
}

func oeError(res C.oe_result_t) error {
	return errors.New(C.GoString(C.oe_result_str(res)))
}
