// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ertcrypto

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mock struct{}

func (mock) GetUniqueSealKey() (key, keyInfo []byte, err error) {
	return []byte("1234567890123456"), []byte("unique"), nil
}
func (mock) GetProductSealKey() (key, keyInfo []byte, err error) {
	return []byte("2345678901234567"), []byte("product"), nil
}
func (mock) GetSealKey(keyInfo []byte) ([]byte, error) {
	switch string(keyInfo) {
	case "unique":
		return []byte("1234567890123456"), nil
	case "product":
		return []byte("2345678901234567"), nil
	}
	return nil, errors.New("unknown keyInfo")
}

func init() {
	sealer = mock{}
}

func TestEncryptAndDecrypt(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Test parameters
	encryptionKey := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	testString := "Edgeless"

	// Encrypt text
	ciphertext, err := Encrypt([]byte(testString), encryptionKey)
	require.NoError(err)

	// Decrypt text
	plaintext, err := Decrypt(ciphertext, encryptionKey)
	require.NoError(err)
	assert.EqualValues(testString, plaintext)
}

func TestSealAndUnseal(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	testString := "Edgeless"

	ciphertext, err := SealWithUniqueKey([]byte(testString))
	require.NoError(err)
	plaintext, err := Unseal(ciphertext)
	require.NoError(err)
	assert.EqualValues(testString, plaintext)

	ciphertext, err = SealWithProductKey([]byte(testString))
	require.NoError(err)
	plaintext, err = Unseal(ciphertext)
	require.NoError(err)
	assert.EqualValues(testString, plaintext)
}

func TestCorruptedUnseal(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	testString := "Edgeless"

	// Check for error if the given ciphertext is nil
	_, err := Unseal(nil)
	assert.Error(err)

	// Check for error if the given ciphertext is too short
	_, err = Unseal([]byte{0, 1, 2})
	assert.Error(err)

	// Check for error if the embedded length is 0
	_, err = Unseal([]byte{0, 0, 0, 0})
	assert.Error(err)

	// Check for error if the embedded ciphertext is invalid (and specifically, if nonce slicing is not out of bounds)
	_, err = Unseal([]byte{4, 0, 0, 0, 'i', 'n', 'f', 'o'})
	assert.Error(err)

	// Check for error if we go out of bounds with an invalid key info length
	ciphertext, err := SealWithUniqueKey([]byte(testString))
	require.NoError(err)

	// Flip two size bits and watch the length go boom :)
	ciphertext[0] = 0xff
	ciphertext[1] = 0xff

	// But hopefully, we catched that!
	_, err = Unseal(ciphertext)
	assert.Error(err)
}

func TestInternalSeal(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	// Test parameters
	sealKey := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	keyInfo := []byte{15, 14, 13, 12, 11, 10, 9, 8, 7, 42, 42, 42}
	testString := "Edgeless"

	// Seal with the given parameters
	sealedText, err := seal([]byte(testString), sealKey, keyInfo)
	require.NoError(err)

	// Check structure of the sealed data
	keyInfoLength := sealedText[:4]
	actualKeyInfoLength := binary.LittleEndian.Uint32(keyInfoLength)
	assert.EqualValues(len(keyInfo), actualKeyInfoLength)

	// Check if keyInfo was written correctly and is at the correct position
	actualKeyInfo := sealedText[4 : 4+len(keyInfo)]
	assert.Equal(keyInfo, actualKeyInfo)

	// Check if ciphertext can be decrypted correctly
	ciphertext := sealedText[4+len(keyInfo):]
	plaintext, err := Decrypt(ciphertext, sealKey)
	require.NoError(err)
	assert.EqualValues(testString, plaintext)
}
