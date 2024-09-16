// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package attestation

// #include "claim.h"
// #include "sgx_evidence.h"
import "C"

import (
	"errors"
	"fmt"
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
	var reportSGX SGXClaims
	var reportSGXOptional SGXOptional
	var claimCountSGXrequired = 0
	var claimCountSGXoptional = 0

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
		case C.OE_CLAIM_UEID:
			// The UEID is prefixed with a type which is currently always OE_UEID_TYPE_RAND for SGX
			claimUEID := claimBytes(claim)
			if len(claimUEID) > 0 && claimUEID[0] != C.OE_UEID_TYPE_RAND {
				return Report{}, errors.New("Expected UEID of type OE_UEID_TYPE_RAND")
			}
			report.UEID = claimUEID
			// SGX Required claims
		case C.OE_CLAIM_SGX_PF_GP_EXINFO_ENABLED:
			reportSGX.SGXRequired.PfGpExinfoEnabled = claimBool(claim)
			claimCountSGXrequired++
		case C.OE_CLAIM_SGX_ISV_EXTENDED_PRODUCT_ID:
			reportSGX.SGXRequired.ISVExtendedProductID = claimBytes(claim)
			claimCountSGXrequired++
		case C.OE_CLAIM_SGX_IS_MODE64BIT:
			reportSGX.SGXRequired.IsMode64Bit = claimBool(claim)
			claimCountSGXrequired++
		case C.OE_CLAIM_SGX_HAS_PROVISION_KEY:
			reportSGX.SGXRequired.HasProvisionKey = claimBool(claim)
			claimCountSGXrequired++
		case C.OE_CLAIM_SGX_HAS_EINITTOKEN_KEY:
			reportSGX.SGXRequired.HasEINITTokenKey = claimBool(claim)
			claimCountSGXrequired++
		case C.OE_CLAIM_SGX_USES_KSS:
			reportSGX.SGXRequired.UsesKSS = claimBool(claim)
			claimCountSGXrequired++
		case C.OE_CLAIM_SGX_CONFIG_ID:
			reportSGX.SGXRequired.ConfigID = claimBytes(claim)
			claimCountSGXrequired++
		case C.OE_CLAIM_SGX_CONFIG_SVN:
			reportSGX.SGXRequired.ConfigSVN = claimBytes(claim)
			claimCountSGXrequired++
		case C.OE_CLAIM_SGX_ISV_FAMILY_ID:
			reportSGX.SGXRequired.ISVFamilyID = claimBytes(claim)
			claimCountSGXrequired++
		case C.OE_CLAIM_SGX_CPU_SVN:
			reportSGX.SGXRequired.CPUSVN = claimBytes(claim)
			claimCountSGXrequired++
			//SGX optional claims
		case C.OE_CLAIM_SGX_TCB_INFO:
			reportSGXOptional.TCBInfo = claimBytes(claim)
			claimCountSGXoptional++
		case C.OE_CLAIM_SGX_TCB_ISSUER_CHAIN:
			reportSGXOptional.TCBIssuerChain = claimBytes(claim)
			claimCountSGXoptional++
		case C.OE_CLAIM_SGX_PCK_CRL:
			reportSGXOptional.PCKCRL = claimBytes(claim)
			claimCountSGXoptional++
		case C.OE_CLAIM_SGX_ROOT_CA_CRL:
			reportSGXOptional.RootCACRL = claimBytes(claim)
			claimCountSGXoptional++
		case C.OE_CLAIM_SGX_CRL_ISSUER_CHAIN:
			reportSGXOptional.CRLIssuerChain = claimBytes(claim)
			claimCountSGXoptional++
		case C.OE_CLAIM_SGX_QE_ID_INFO:
			reportSGXOptional.QEIDInfo = claimBytes(claim)
			claimCountSGXoptional++
		case C.OE_CLAIM_SGX_QE_ID_ISSUER_CHAIN:
			reportSGXOptional.QEIDIssuerChain = claimBytes(claim)
			claimCountSGXoptional++
		case C.OE_CLAIM_SGX_PCE_SVN:
			reportSGXOptional.PCESVN = claimBytes(claim)
			claimCountSGXoptional++
		}

	}
	if claimCountSGXrequired > 0 && claimCountSGXrequired != C.OE_SGX_REQUIRED_CLAIMS_COUNT {
		return Report{}, fmt.Errorf("some required SGX claims are missing. Only got: %d, expected: %d", claimCountSGXrequired, C.OE_SGX_REQUIRED_CLAIMS_COUNT)
	}

	if claimCountSGXoptional > C.OE_SGX_OPTIONAL_CLAIMS_COUNT {
		return Report{}, fmt.Errorf("optional SGX claims are too many. Got: %d, expected maximum: %d", claimCountSGXoptional, C.OE_SGX_OPTIONAL_CLAIMS_COUNT)
	}

	if !hasAttributes {
		return Report{}, errors.New("missing attributes in report claims")
	}

	if claimCountSGXoptional > 0 {
		reportSGX.SGXOptional = &reportSGXOptional
	}

	if claimCountSGXrequired > 0 {
		report.SGXClaims = &reportSGX
	}

	return report, nil
}

func claimUint(claim C.oe_claim_t) uint {
	if claim.value_size < 4 {
		return 0
	}
	return uint(*(*C.uint32_t)(unsafe.Pointer(claim.value)))
}

func claimBool(claim C.oe_claim_t) bool {
	return bool(*(*C._Bool)(unsafe.Pointer(claim.value)))
}

func claimBytes(claim C.oe_claim_t) []byte {
	return C.GoBytes(unsafe.Pointer(claim.value), C.int(claim.value_size))
}
