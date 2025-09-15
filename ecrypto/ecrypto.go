// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ecrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"

	"github.com/edgelesssys/ego/enclave"
)

const keyInfoLengthLength = 4

var sealer interface {
	GetUniqueSealKey() (key, keyInfo []byte, err error)
	GetProductSealKey() (key, keyInfo []byte, err error)
	GetSealKey(keyInfo []byte) ([]byte, error)
} = enclaveSealer{}

type enclaveSealer struct{}

func (enclaveSealer) GetUniqueSealKey() (key, keyInfo []byte, err error) {
	return enclave.GetRandomUniqueSealKey()
}

func (enclaveSealer) GetProductSealKey() (key, keyInfo []byte, err error) {
	return enclave.GetRandomProductSealKey()
}

func (enclaveSealer) GetSealKey(keyInfo []byte) ([]byte, error) {
	return enclave.GetSealKey(keyInfo)
}

// Encrypt encrypts a given plaintext with a supplied key using AES-GCM.
//
// Optionally pass additionalData to be authenticated.
func Encrypt(plaintext []byte, key []byte, additionalData []byte) ([]byte, error) {
	// Get cipher object with key
	aesgcm, err := getCipher(key)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt data
	return aesgcm.Seal(nonce, nonce, plaintext, additionalData), nil
}

// Decrypt decrypts a ciphertext produced by Encrypt.
//
// The additionalData must match the value passed to Encrypt.
func Decrypt(ciphertext []byte, key []byte, additionalData []byte) ([]byte, error) {
	// Get cipher object with key
	aesgcm, err := getCipher(key)
	if err != nil {
		return nil, err
	}

	// Split ciphertext into nonce and actual data
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) <= nonceSize {
		return nil, errors.New("ciphertext is too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt data
	return aesgcm.Open(nil, nonce, ciphertext, additionalData)
}

// SealWithUniqueKey encrypts a given plaintext with a key derived from a measurement of the enclave.
//
// Optionally pass additionalData to be authenticated.
//
// Ciphertexts can't be decrypted if the UniqueID of the enclave changes. If you want
// to be able to decrypt ciphertext across enclave versions, use SealWithProductKey.
func SealWithUniqueKey(plaintext []byte, additionalData []byte) ([]byte, error) {
	sealKey, keyInfo, err := sealer.GetUniqueSealKey()
	if err != nil {
		return nil, err
	}

	return seal(plaintext, sealKey, keyInfo, additionalData)
}

// SealWithProductKey encrypts a given plaintext with a key derived from the signer and product id of the enclave.
//
// Optionally pass additionalData to be authenticated.
func SealWithProductKey(plaintext []byte, additionalData []byte) ([]byte, error) {
	sealKey, keyInfo, err := sealer.GetProductSealKey()
	if err != nil {
		return nil, err
	}

	return seal(plaintext, sealKey, keyInfo, additionalData)
}

// Unseal decrypts a ciphertext produced by SealWithUniqueKey or SealWithProductKey.
//
// The additionalData must match the value passed to Encrypt.
func Unseal(ciphertext []byte, additionalData []byte) ([]byte, error) {
	keyInfo, ciphertext, err := splitCiphertext(ciphertext)
	if err != nil {
		return nil, err
	}

	sealKey, err := sealer.GetSealKey(keyInfo)
	if err != nil {
		return nil, err
	}

	return Decrypt(ciphertext, sealKey, additionalData)
}

// SealWithUniqueKey256 encrypts a given plaintext with a key derived from a measurement of the enclave.
//
// It constructs a 256-bit key by concatenating the results of two EGETKEY calls with different random KeyIDs.
//
// Optionally pass additionalData to be authenticated.
//
// Ciphertexts can't be decrypted if the UniqueID of the enclave changes. If you want
// to be able to decrypt ciphertext across enclave versions, use SealWithProductKey256.
func SealWithUniqueKey256(plaintext []byte, additionalData []byte) ([]byte, error) {
	sealKey1, keyInfo1, err := sealer.GetUniqueSealKey()
	if err != nil {
		return nil, err
	}
	sealKey2, keyInfo2, err := sealer.GetUniqueSealKey()
	if err != nil {
		return nil, err
	}
	sealKey := append(sealKey1, sealKey2...)
	keyInfo := append(keyInfo1, keyInfo2...)

	return seal(plaintext, sealKey, keyInfo, additionalData)
}

// SealWithProductKey256 encrypts a given plaintext with a key derived from the signer and product id of the enclave.
//
// It constructs a 256-bit key by concatenating the results of two EGETKEY calls with different random KeyIDs.
//
// Optionally pass additionalData to be authenticated.
func SealWithProductKey256(plaintext []byte, additionalData []byte) ([]byte, error) {
	sealKey1, keyInfo1, err := sealer.GetProductSealKey()
	if err != nil {
		return nil, err
	}
	sealKey2, keyInfo2, err := sealer.GetProductSealKey()
	if err != nil {
		return nil, err
	}
	sealKey := append(sealKey1, sealKey2...)
	keyInfo := append(keyInfo1, keyInfo2...)

	return seal(plaintext, sealKey, keyInfo, additionalData)
}

// Unseal256 decrypts a ciphertext produced by SealWithUniqueKey256 or SealWithProductKey256.
//
// The additionalData must match the value passed to Encrypt.
func Unseal256(ciphertext []byte, additionalData []byte) ([]byte, error) {
	keyInfo, ciphertext, err := splitCiphertext(ciphertext)
	if err != nil {
		return nil, err
	}

	length := len(keyInfo) / 2
	sealKey1, err := sealer.GetSealKey(keyInfo[:length])
	if err != nil {
		return nil, err
	}
	sealKey2, err := sealer.GetSealKey(keyInfo[length:])
	if err != nil {
		return nil, err
	}
	sealKey := append(sealKey1, sealKey2...)

	return Decrypt(ciphertext, sealKey, additionalData)
}

func getCipher(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func seal(plaintext []byte, sealKey []byte, keyInfo []byte, additionalData []byte) ([]byte, error) {
	// Encrypt plaintext with the given seal key
	ciphertext, err := Encrypt(plaintext, sealKey, additionalData)
	if err != nil {
		return nil, err
	}

	// Encode keyInfo and its length and append it in front of the ciphertext
	// Use a fixed size length so we can properly extract the length from the ciphertext when unsealing
	keyInfoLength := make([]byte, keyInfoLengthLength)
	binary.LittleEndian.PutUint32(keyInfoLength, uint32(len(keyInfo)))
	keyInfoEncoded := append(keyInfoLength, keyInfo...)

	return append(keyInfoEncoded, ciphertext...), nil
}

func splitCiphertext(ciphertext []byte) (keyInfo, actualCiphertext []byte, err error) {
	// pop key info length from ciphertext front
	if len(ciphertext) <= keyInfoLengthLength {
		return nil, nil, errors.New("ciphertext is too short")
	}
	keyInfoLength := binary.LittleEndian.Uint32(ciphertext[:keyInfoLengthLength])
	ciphertext = ciphertext[keyInfoLengthLength:]

	// split ciphertext into key info and actual data
	if !(0 < keyInfoLength && int(keyInfoLength) < len(ciphertext)) {
		return nil, nil, errors.New("ciphertext contains invalid key info length")
	}
	return ciphertext[:keyInfoLength], ciphertext[keyInfoLength:], nil
}
