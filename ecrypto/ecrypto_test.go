// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ecrypto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubSealer struct{}

func (stubSealer) GetUniqueSealKey() (key, keyInfo []byte, err error) {
	return []byte("1234567890123456"), []byte("unique"), nil
}

func (stubSealer) GetProductSealKey() (key, keyInfo []byte, err error) {
	return []byte("2345678901234567"), []byte("product"), nil
}

func (stubSealer) GetSealKey(keyInfo []byte) ([]byte, error) {
	switch string(keyInfo) {
	case "unique":
		return []byte("1234567890123456"), nil
	case "product":
		return []byte("2345678901234567"), nil
	}
	return nil, errors.New("unknown keyInfo")
}

func init() {
	sealer = stubSealer{}
}

func TestEncryptDecrypt(t *testing.T) {
	testKey := []byte("0123456789012345")
	testCases := map[string]struct {
		plaintext      string
		key            []byte
		additionalData []byte
		expectErr      bool
	}{
		"basic": {
			plaintext: "foo",
			key:       testKey,
		},
		"empty plaintext": {
			key: testKey,
		},
		"long plaintext": {
			plaintext: strings.Repeat("Edgeless Systems", 100),
			key:       testKey,
		},
		"nil key is invalid": {
			plaintext: "foo",
			key:       nil,
			expectErr: true,
		},
		"empty key is invalid": {
			plaintext: "foo",
			key:       []byte{},
			expectErr: true,
		},
		"15 byte key is invalid": {
			plaintext: "foo",
			key:       make([]byte, 15),
			expectErr: true,
		},
		"all zero key is valid": {
			plaintext: "foo",
			key:       make([]byte, 16),
		},
		"20 byte key is invalid": {
			plaintext: "foo",
			key:       make([]byte, 20),
			expectErr: true,
		},
		"24 byte key is valid": {
			plaintext: "foo",
			key:       make([]byte, 24),
		},
		"additional data": {
			plaintext:      "foo",
			key:            testKey,
			additionalData: []byte{2, 3, 4},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			plaintext := []byte(tc.plaintext)
			ciphertext, err := Encrypt(plaintext, tc.key, tc.additionalData)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Greater(len(ciphertext), len(plaintext))
			if tc.plaintext != "" {
				assert.False(bytes.Contains(ciphertext, plaintext))
			}

			decryptedPlaintext, err := Decrypt(ciphertext, tc.key, tc.additionalData)
			require.NoError(err)
			if tc.plaintext == "" {
				assert.Empty(decryptedPlaintext)
			} else {
				assert.Equal(plaintext, decryptedPlaintext)
			}
		})
	}
}

func TestEncryptIsRandom(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	plaintext := []byte("foo")
	key1 := []byte("0123456789012345")
	key2 := []byte("0123456789012346")

	ciphertext1, err := Encrypt(plaintext, key1, nil)
	require.NoError(err)
	ciphertext2, err := Encrypt(plaintext, key2, nil)
	require.NoError(err)
	ciphertext1a, err := Encrypt(plaintext, key1, nil)
	require.NoError(err)

	assert.NotEqual(ciphertext2, ciphertext1)
	assert.NotEqual(ciphertext1a, ciphertext1)
}

func TestDecryptError(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	key1 := []byte("0123456789012345")
	key2 := []byte("0123456789012346")

	ciphertext, err := Encrypt([]byte("foo"), key1, nil)
	require.NoError(err)

	_, err = Decrypt(ciphertext, key2, nil)
	assert.Error(err)

	ciphertext[22] ^= 1
	_, err = Decrypt(ciphertext, key1, nil)
	assert.Error(err)
}

func TestDecryptAdditionalDataError(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	key := []byte("0123456789012345")

	ciphertext, err := Encrypt([]byte("foo"), key, []byte{2, 3, 4})
	require.NoError(err)

	_, err = Decrypt(ciphertext, key, nil)
	assert.Error(err)

	_, err = Decrypt(ciphertext, key, []byte{2, 3, 5})
	assert.Error(err)
}

func TestSealUnseal(t *testing.T) {
	testCases := map[string]struct {
		seal           func(plaintext, additionalData []byte) ([]byte, error)
		plaintext      string
		additionalData []byte
	}{
		"unique: basic": {
			seal:      SealWithUniqueKey,
			plaintext: "foo",
		},
		"unique: empty plaintext": {
			seal: SealWithUniqueKey,
		},
		"unique: long plaintext": {
			seal:      SealWithUniqueKey,
			plaintext: strings.Repeat("Edgeless Systems", 100),
		},
		"unique: additional data": {
			seal:           SealWithUniqueKey,
			plaintext:      "foo",
			additionalData: []byte{2, 3, 4},
		},
		"product: basic": {
			seal:      SealWithProductKey,
			plaintext: "foo",
		},
		"product: empty plaintext": {
			seal: SealWithProductKey,
		},
		"product: long plaintext": {
			seal:      SealWithProductKey,
			plaintext: strings.Repeat("Edgeless Systems", 100),
		},
		"product: additional data": {
			seal:           SealWithProductKey,
			plaintext:      "foo",
			additionalData: []byte{2, 3, 4},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			plaintext := []byte(tc.plaintext)
			ciphertext, err := tc.seal(plaintext, tc.additionalData)
			require.NoError(err)

			assert.Greater(len(ciphertext), len(plaintext))
			if tc.plaintext != "" {
				assert.False(bytes.Contains(ciphertext, plaintext))
			}

			decryptedPlaintext, err := Unseal(ciphertext, tc.additionalData)
			require.NoError(err)
			if tc.plaintext == "" {
				assert.Empty(decryptedPlaintext)
			} else {
				assert.Equal(plaintext, decryptedPlaintext)
			}
		})
	}
}

func TestUnsealError(t *testing.T) {
	testCases := map[string]struct {
		ciphertext []byte
	}{
		"nil":                               {nil},
		"empty":                             {[]byte{}},
		"shorter than keyInfoLength length": {[]byte{2, 3, 4}},
		"keyInfoLength=0":                   {[]byte{0, 0, 0, 0}},
		"keyInfoLength=1":                   {[]byte{1, 0, 0, 0}},
		"keyInfoLength=2":                   {[]byte{2, 0, 0, 0}},
		"keyInfoLength=max":                 {[]byte{255, 255, 255, 255}},
		"keyInfo without ciphertext":        {[]byte{4, 0, 0, 0, 'i', 'n', 'f', 'o'}},
		"too short keyInfo":                 {[]byte{5, 0, 0, 0, 'i', 'n', 'f', 'o'}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			_, err := Unseal(tc.ciphertext, nil)
			assert.Error(err)
		})
	}
}

func TestUnsealAdditionalDataError(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	ciphertext, err := SealWithUniqueKey([]byte("foo"), []byte{2, 3, 4})
	require.NoError(err)

	_, err = Unseal(ciphertext, nil)
	assert.Error(err)

	_, err = Unseal(ciphertext, []byte{2, 3, 5})
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
	sealedText, err := seal([]byte(testString), sealKey, keyInfo, nil)
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
	plaintext, err := Decrypt(ciphertext, sealKey, nil)
	require.NoError(err)
	assert.EqualValues(testString, plaintext)
}
