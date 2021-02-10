// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package ertenclave provides functionality for Go enclaves like remote attestation and sealing.
package ertenclave

// #include "structs.h"
import "C"

import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/edgelesssys/ertgolib/ert"
)

const SYS_get_remote_report = 1000
const SYS_free_report = 1001
const SYS_verify_report = 1002
const SYS_get_seal_key = 1003
const SYS_free_seal_key = 1004
const SYS_get_seal_key_by_policy = 1005
const SYS_result_str = 1006

// GetRemoteReport gets a report signed by the enclave platform for use in remote attestation.
//
// The report shall contain the data given by the reportData parameter.
func GetRemoteReport(reportData []byte) ([]byte, error) {
	var report *C.uint8_t
	var reportSize C.size_t

	res, _, errno := syscall.Syscall6(
		SYS_get_remote_report,
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
	syscall.Syscall(SYS_free_report, uintptr(unsafe.Pointer(report)), 0, 0)
	return result, nil
}

// VerifyRemoteReport verifies the integrity of the remote report and its signature.
//
// This function verifies that the report signature is valid. It
// verifies that the signing authority is rooted to a trusted authority
// such as the enclave platform manufacturer.
//
// Returns the parsed report if the signature is valid.
// Returns an error if the signature is invalid.
func VerifyRemoteReport(reportBytes []byte) (ert.Report, error) {
	var report C.oe_report_t

	res, _, errno := syscall.Syscall(
		SYS_verify_report,
		uintptr(unsafe.Pointer(&reportBytes[0])),
		uintptr(len(reportBytes)),
		uintptr(unsafe.Pointer(&report)),
	)
	if err := oeError(errno, res); err != nil {
		return ert.Report{}, err
	}

	if (report.identity.attributes & C.OE_REPORT_ATTRIBUTES_REMOTE) == 0 {
		return ert.Report{}, errors.New("OE_UNSUPPORTED")
	}

	return ert.Report{
		Data:            C.GoBytes(unsafe.Pointer(report.report_data), C.int(report.report_data_size)),
		SecurityVersion: uint(report.identity.security_version),
		Debug:           (report.identity.attributes & C.OE_REPORT_ATTRIBUTES_DEBUG) != 0,
		UniqueID:        C.GoBytes(unsafe.Pointer(&report.identity.unique_id[0]), C.OE_UNIQUE_ID_SIZE),
		SignerID:        C.GoBytes(unsafe.Pointer(&report.identity.signer_id[0]), C.OE_SIGNER_ID_SIZE),
		ProductID:       C.GoBytes(unsafe.Pointer(&report.identity.product_id[0]), C.OE_PRODUCT_ID_SIZE),
	}, nil
}

// GetUniqueSealKey gets a key derived from a measurement of the enclave.
//
// keyInfo can be used to retrieve the same key later, on a newer security version.
func GetUniqueSealKey() (key, keyInfo []byte, err error) {
	return getSealKeyByPolicy(C.OE_SEAL_POLICY_UNIQUE)
}

// GetProductSealKey gets a key derived from the signer and product id of the enclave.
//
// keyInfo can be used to retrieve the same key later, on a newer security version.
func GetProductSealKey() (key, keyInfo []byte, err error) {
	return getSealKeyByPolicy(C.OE_SEAL_POLICY_PRODUCT)
}

// GetSealKey gets a key from the enclave platform using existing key information.
func GetSealKey(keyInfo []byte) ([]byte, error) {
	var keyBuffer *C.uint8_t
	var keySize C.size_t

	res, _, errno := syscall.Syscall6(
		SYS_get_seal_key,
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
	syscall.Syscall(SYS_free_seal_key, uintptr(unsafe.Pointer(keyBuffer)), 0, 0)
	return key, nil
}

func getSealKeyByPolicy(sealPolicy uintptr) (key, keyInfo []byte, err error) {
	var keyBuffer, keyInfoBuffer *C.uint8_t
	var keySize, keyInfoSize C.size_t

	res, _, errno := syscall.Syscall6(
		SYS_get_seal_key_by_policy,
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
	syscall.Syscall(
		SYS_free_seal_key,
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

	resStr, _, errno := syscall.Syscall(SYS_result_str, res, 0, 0)
	if errno != 0 {
		return errno
	}
	return errors.New(C.GoString((*C.char)(unsafe.Pointer(resStr))))
}
