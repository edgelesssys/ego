// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package attestation

// #include "claim.h"
import "C"

import (
	"errors"
	"unsafe"

	"github.com/edgelesssys/ego/attestation/tcbstatus"
)

func ParseClaims(claims uintptr, claimsLength uintptr) (Report, error) {
	// https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices
	return parseClaims((*[1 << 28]C.oe_claim_t)(unsafe.Pointer(claims))[:claimsLength:claimsLength])
}

func parseClaims(claims []C.oe_claim_t) (Report, error) {
	report := Report{TCBStatus: tcbstatus.Unknown}
	hasAttributes := false

	for _, claim := range claims {
		switch C.GoString(claim.name) {
		case C.OE_CLAIM_SECURITY_VERSION:
			report.SecurityVersion = claimUint(claim)
		case C.OE_CLAIM_ATTRIBUTES:
			hasAttributes = true
			attr := claimUint(claim)
			if (attr & C.OE_EVIDENCE_ATTRIBUTES_SGX_REMOTE) == 0 {
				return Report{}, errors.New("not a remote report")
			}
			report.Debug = (attr & C.OE_EVIDENCE_ATTRIBUTES_SGX_DEBUG) != 0
		case C.OE_CLAIM_UNIQUE_ID:
			report.UniqueID = claimBytes(claim)
		case C.OE_CLAIM_SIGNER_ID:
			report.SignerID = claimBytes(claim)
		case C.OE_CLAIM_PRODUCT_ID:
			report.ProductID = claimBytes(claim)
		case C.OE_CLAIM_TCB_STATUS:
			report.TCBStatus = tcbstatus.Status(claimUint(claim))
		case C.OE_CLAIM_SGX_REPORT_DATA:
			report.Data = claimBytes(claim)
		}
	}

	if !hasAttributes {
		return Report{}, errors.New("missing attributes in report claims")
	}
	return report, nil
}

func claimUint(claim C.oe_claim_t) uint {
	if claim.value_size < 4 {
		return 0
	}
	return uint(*(*C.uint32_t)(unsafe.Pointer(claim.value)))
}

func claimBytes(claim C.oe_claim_t) []byte {
	return C.GoBytes(unsafe.Pointer(claim.value), C.int(claim.value_size))
}
