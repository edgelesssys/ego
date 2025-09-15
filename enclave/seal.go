package enclave

import (
	"crypto/rand"
	"encoding/binary"
)

const (
	sgxKeyselectSeal      = 4
	sgxKeypolicyMRENCLAVE = 1
	sgxKeypolicyMRSIGNER  = 2

	// https://github.com/intel/linux-sgx/blob/sgx_2.3/common/inc/sgx_report.h
	offsetReportCPUSVN = 0
	offsetReportISVSVN = 258
)

// GetSealKeyID gets a unique ID derived from the CPU's root seal key.
// The ID also depends on the ProductID and Debug flag of the enclave.
func GetSealKeyID() ([]byte, error) {
	// Leaving all other fields 0 means only ProductID (but not unique or signer id) of the enclave
	// is used for derivation. The key isn't secret because everyone can create an enclave with the
	// same ProductID that prints the key. So we can directly use it as an ID for the CPU.
	keyRequest, err := sgxKeyRequest{KeyName: sgxKeyselectSeal}.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return GetSealKey(keyRequest)
}

// GetRandomUniqueSealKey gets a key derived from a measurement of the enclave.
//
// keyInfo can be used to retrieve the same key later.
//
// This key will change if the UniqueID of the enclave changes. If you want
// the key to be the same across enclave versions, use GetRandomProductSealKey.
func GetRandomUniqueSealKey() (key, keyInfo []byte, err error) {
	return getSealKeyByPolicy(sgxKeypolicyMRENCLAVE, true)
}

// GetRandomProductSealKey gets a key derived from the signer and product id of the enclave.
//
// keyInfo can be used to retrieve the same key later.
func GetRandomProductSealKey() (key, keyInfo []byte, err error) {
	return getSealKeyByPolicy(sgxKeypolicyMRSIGNER, true)
}

// GetUniqueSealKey gets a key derived from a measurement of the enclave.
//
// Deprecated: use GetRandomUniqueSealKey
//
// keyInfo can be used to retrieve the same key later, on a newer CPU security version.
//
// This key will change if the UniqueID of the enclave changes. If you want
// the key to be the same across enclave versions, use GetProductSealKey.
func GetUniqueSealKey() (key, keyInfo []byte, err error) {
	return getSealKeyByPolicy(sgxKeypolicyMRENCLAVE, false)
}

// GetProductSealKey gets a key derived from the signer and product id of the enclave.
//
// Deprecated: use GetRandomProductSealKey
//
// keyInfo can be used to retrieve the same key later, on a newer CPU security version.
func GetProductSealKey() (key, keyInfo []byte, err error) {
	return getSealKeyByPolicy(sgxKeypolicyMRSIGNER, false)
}

func getSealKeyByPolicy(sealPolicy uint16, random bool) (key, keyInfo []byte, err error) {
	// https://github.com/openenclave/openenclave/blob/v0.19.13/enclave/core/sgx/keys.c#L191
	report, err := GetLocalReport(nil, nil)
	if err != nil {
		return nil, nil, err
	}
	report = report[16:] // skip OE header
	req := sgxKeyRequest{
		KeyName:   sgxKeyselectSeal,
		KeyPolicy: sealPolicy,
		ISVSVN:    binary.LittleEndian.Uint16(report[offsetReportISVSVN:]),
	}
	copy(req.CPUSVN[:], report[offsetReportCPUSVN:])
	req.Flags, req.XFRM, req.MiscMask = getSealMasks()

	// https://github.com/openenclave/openenclave/issues/4665
	if random {
		if _, err := rand.Read(req.KeyID[:]); err != nil {
			return nil, nil, err
		}
	}

	keyInfo, err = req.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	key, err = GetSealKey(keyInfo)
	return key, keyInfo, err
}

// https://github.com/intel/linux-sgx/blob/sgx_2.3/common/inc/sgx_key.h
type sgxKeyRequest struct {
	KeyName   uint16
	KeyPolicy uint16
	ISVSVN    uint16
	reserved1 uint16
	CPUSVN    [16]byte
	Flags     uint64
	XFRM      uint64
	KeyID     [32]byte
	MiscMask  uint32
}

func (k sgxKeyRequest) MarshalBinary() ([]byte, error) {
	bin := make([]byte, 0, 512)
	bin = binary.LittleEndian.AppendUint16(bin, k.KeyName)
	bin = binary.LittleEndian.AppendUint16(bin, k.KeyPolicy)
	bin = binary.LittleEndian.AppendUint16(bin, k.ISVSVN)
	bin = binary.LittleEndian.AppendUint16(bin, k.reserved1)
	bin = append(bin, k.CPUSVN[:]...)
	bin = binary.LittleEndian.AppendUint64(bin, k.Flags)
	bin = binary.LittleEndian.AppendUint64(bin, k.XFRM)
	bin = append(bin, k.KeyID[:]...)
	bin = binary.LittleEndian.AppendUint32(bin, k.MiscMask)
	return bin[:cap(bin)], nil
}
