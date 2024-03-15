// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package enclave

// #include "ert.h"
import "C"

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/edgelesssys/ego/attestation"
	"github.com/edgelesssys/ego/attestation/tcbstatus"
	internal "github.com/edgelesssys/ego/internal/attestation"
)

const (
	sysGetRemoteReport    = 1000
	sysFreeReport         = 1001
	sysVerifyReport       = 1002
	sysGetSealKey         = 1003
	sysFreeSealKey        = 1004
	sysGetSealKeyByPolicy = 1005
	sysResultStr          = 1006
	sysVerifyEvidence     = 1007
	sysFreeClaims         = 1008
	sysGetLocalReport     = 1009
)

const maxReportData = 64

var errReportDataTooLarge = errors.New("reportData too large")

// GetRemoteReport gets a report signed by the enclave platform for use in remote attestation.
//
// The report shall contain the data given by the reportData parameter. The report can only
// hold a maximum of 64 byte reportData. Use a hash value of your data as reportData if your
// data exceeds this limit.
//
// If reportData is less than 64 bytes, it will be padded with zero bytes.
func GetRemoteReport(reportData []byte) ([]byte, error) {
	if len(reportData) > maxReportData {
		return nil, errReportDataTooLarge
	}

	var report *C.uint8_t
	var reportSize C.size_t

	res, _, errno := syscall.Syscall6(
		sysGetRemoteReport,
		uintptr(unsafe.Pointer(&reportData[0])),
		uintptr(len(reportData)),
		0,
		0,
		uintptr(unsafe.Pointer(&report)),
		uintptr(unsafe.Pointer(&reportSize)),
	)
	if err := oeError(errno, res); err != nil {
		return nil, err
	}

	result := C.GoBytes(unsafe.Pointer(report), C.int(reportSize))
	_, _, _ = syscall.Syscall(sysFreeReport, uintptr(unsafe.Pointer(report)), 0, 0)
	return result, nil
}

// VerifyRemoteReport verifies the integrity of the remote report and its signature.
//
// This function verifies that the report signature is valid. It
// verifies that the signing authority is rooted to a trusted authority
// such as the enclave platform manufacturer.
//
// The caller must verify the returned report's content.
func VerifyRemoteReport(reportBytes []byte) (attestation.Report, error) {
	if len(reportBytes) <= 0 {
		return attestation.Report{}, attestation.ErrEmptyReport
	}

	var claims, claimsLength uintptr

	res, _, errno := syscall.Syscall6(
		sysVerifyEvidence,
		uintptr(unsafe.Pointer(&reportBytes[0])),
		uintptr(len(reportBytes)),
		uintptr(unsafe.Pointer(&claims)),
		uintptr(unsafe.Pointer(&claimsLength)),
		0, 0,
	)

	var verifyErr error
	if err := oeError(errno, res); err != nil {
		if err.Error() != "OE_TCB_LEVEL_INVALID" {
			return attestation.Report{}, err
		}
		verifyErr = attestation.ErrTCBLevelInvalid
	}

	defer func() { _, _, _ = syscall.Syscall(sysFreeClaims, claims, claimsLength, 0) }()

	report, err := internal.ParseClaims(claims, claimsLength)
	if err != nil {
		return attestation.Report{}, err
	}
	return attestation.Report{
		Data:            report.Data,
		SecurityVersion: report.SecurityVersion,
		Debug:           report.Debug,
		UniqueID:        report.UniqueID,
		SignerID:        report.SignerID,
		ProductID:       report.ProductID,
		TCBStatus:       report.TCBStatus,
	}, verifyErr
}

// GetLocalReport gets a report signed by the enclave platform for use in local attestation.
//
// The report shall contain the data given by the reportData parameter. The report can only
// hold a maximum of 64 byte reportData. Use a hash value of your data as reportData if your
// data exceeds this limit.
//
// If reportData is less than 64 bytes, it will be padded with zero bytes.
//
// The report can only be verified by the enclave identified by targetReport. So you must
// first get a report from the target enclave. This report is allowed to be empty, i.e.,
// obtained by `GetLocalReport(nil, nil)`.
func GetLocalReport(reportData []byte, targetReport []byte) ([]byte, error) {
	if len(reportData) > maxReportData {
		return nil, errReportDataTooLarge
	}

	var report *C.uint8_t
	var reportSize C.size_t

	res, _, errno := syscall.Syscall6(
		sysGetLocalReport,
		getBytesPointer(reportData),
		uintptr(len(reportData)),
		getBytesPointer(targetReport),
		uintptr(len(targetReport)),
		uintptr(unsafe.Pointer(&report)),
		uintptr(unsafe.Pointer(&reportSize)),
	)
	if err := oeError(errno, res); err != nil {
		return nil, err
	}

	result := C.GoBytes(unsafe.Pointer(report), C.int(reportSize))
	_, _, _ = syscall.Syscall(sysFreeReport, uintptr(unsafe.Pointer(report)), 0, 0)
	return result, nil
}

// VerifyLocalReport verifies the integrity of the local report and its signature.
//
// This function verifies that the report signature is valid. It
// verifies that it is correctly signed by the enclave platform.
//
// The caller must verify the returned report's content.
func VerifyLocalReport(reportBytes []byte) (attestation.Report, error) {
	if len(reportBytes) <= 0 {
		return attestation.Report{}, attestation.ErrEmptyReport
	}

	var report C.oe_report_t

	res, _, errno := syscall.Syscall(
		sysVerifyReport,
		uintptr(unsafe.Pointer(&reportBytes[0])),
		uintptr(len(reportBytes)),
		uintptr(unsafe.Pointer(&report)),
	)
	if err := oeError(errno, res); err != nil {
		return attestation.Report{}, err
	}

	if (report.identity.attributes & C.OE_REPORT_ATTRIBUTES_REMOTE) != 0 {
		return attestation.Report{}, errors.New("expected a local report, but got a remote report")
	}

	return attestation.Report{
		Data:            C.GoBytes(unsafe.Pointer(report.report_data), C.int(report.report_data_size)),
		SecurityVersion: uint(report.identity.security_version),
		Debug:           (report.identity.attributes & C.OE_REPORT_ATTRIBUTES_DEBUG) != 0,
		UniqueID:        C.GoBytes(unsafe.Pointer(&report.identity.unique_id[0]), C.OE_UNIQUE_ID_SIZE),
		SignerID:        C.GoBytes(unsafe.Pointer(&report.identity.signer_id[0]), C.OE_SIGNER_ID_SIZE),
		ProductID:       C.GoBytes(unsafe.Pointer(&report.identity.product_id[0]), C.OE_PRODUCT_ID_SIZE),
		TCBStatus:       tcbstatus.Unknown,
	}, nil
}

// GetUniqueSealKey gets a key derived from a measurement of the enclave.
//
// keyInfo can be used to retrieve the same key later, on a newer CPU security version.
//
// This key will change if the UniqueID of the enclave changes. If you want
// the key to be the same across enclave versions, use GetProductSealKey.
func GetUniqueSealKey() (key, keyInfo []byte, err error) {
	return getSealKeyByPolicy(C.OE_SEAL_POLICY_UNIQUE)
}

// GetProductSealKey gets a key derived from the signer and product id of the enclave.
//
// keyInfo can be used to retrieve the same key later, on a newer CPU security version.
func GetProductSealKey() (key, keyInfo []byte, err error) {
	return getSealKeyByPolicy(C.OE_SEAL_POLICY_PRODUCT)
}

// GetSealKey gets a key from the enclave platform using existing key information.
func GetSealKey(keyInfo []byte) ([]byte, error) {
	var keyBuffer *C.uint8_t
	var keySize C.size_t

	res, _, errno := syscall.Syscall6(
		sysGetSealKey,
		uintptr(unsafe.Pointer(&keyInfo[0])),
		uintptr(len(keyInfo)),
		uintptr(unsafe.Pointer(&keyBuffer)),
		uintptr(unsafe.Pointer(&keySize)),
		0,
		0,
	)
	if errno == syscall.ENOSYS {
		return make([]byte, 16), nil
	}
	if err := oeError(errno, res); err != nil {
		return nil, err
	}

	key := C.GoBytes(unsafe.Pointer(keyBuffer), C.int(keySize))
	_, _, _ = syscall.Syscall(sysFreeSealKey, uintptr(unsafe.Pointer(keyBuffer)), 0, 0)
	return key, nil
}

func getSealKeyByPolicy(sealPolicy uintptr) (key, keyInfo []byte, err error) {
	var keyBuffer, keyInfoBuffer *C.uint8_t
	var keySize, keyInfoSize C.size_t

	res, _, errno := syscall.Syscall6(
		sysGetSealKeyByPolicy,
		sealPolicy,
		uintptr(unsafe.Pointer(&keyBuffer)),
		uintptr(unsafe.Pointer(&keySize)),
		uintptr(unsafe.Pointer(&keyInfoBuffer)),
		uintptr(unsafe.Pointer(&keyInfoSize)),
		0,
	)
	if errno == syscall.ENOSYS {
		return make([]byte, 16), []byte("info"), nil
	}
	if err := oeError(errno, res); err != nil {
		return nil, nil, err
	}

	key = C.GoBytes(unsafe.Pointer(keyBuffer), C.int(keySize))
	keyInfo = C.GoBytes(unsafe.Pointer(keyInfoBuffer), C.int(keyInfoSize))
	_, _, _ = syscall.Syscall(
		sysFreeSealKey,
		uintptr(unsafe.Pointer(keyBuffer)),
		uintptr(unsafe.Pointer(keyInfoBuffer)),
		0,
	)
	return
}

func oeError(errno syscall.Errno, res uintptr) error {
	if errno == syscall.ENOSYS {
		return errors.New("OE_UNSUPPORTED")
	}
	if errno != 0 {
		return errno
	}
	if res == 0 {
		return nil
	}

	resStr, _, errno := syscall.Syscall(sysResultStr, res, 0, 0)
	if errno != 0 {
		return errno
	}
	return errors.New(C.GoString((*C.char)(unsafe.Pointer(resStr))))
}

func getBytesPointer(data []byte) uintptr {
	if len(data) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(&data[0]))
}
