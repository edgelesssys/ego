// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

const (
	modulusSize = 384

	offsetSigstruct = 144
	offsetModulus   = 128
	offsetMRENCLAVE = 960
)

func (c *Cli) signeridByKey(path string) (string, error) {
	// get RSA key from PEM file
	pemBytes, err := c.fs.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading key file: %w", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return "", errors.New("failed to decode PEM")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parsing public key: %w", err)
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("expected RSA public key, got %T", key)
	}

	// MRSIGNER is the sha256 of the modulus in little endian
	n := rsaKey.N.FillBytes(make([]byte, modulusSize))
	slices.Reverse(n)
	sum := sha256.Sum256(n)
	return hex.EncodeToString(sum[:]), nil
}

func (c *Cli) signeridByExecutable(path string) (string, error) {
	modulus, err := c.readDataFromELF(path, oeinfoSectionName, offsetSigstruct+offsetModulus, modulusSize)
	if err != nil {
		return "", err
	}
	if bytes.Equal(modulus, make([]byte, modulusSize)) {
		// if enclave is unsigned, return all zeros
		return strings.Repeat("0", 64), nil
	}

	// MRSIGNER is the sha256 of the modulus
	sum := sha256.Sum256(modulus)
	return hex.EncodeToString(sum[:]), nil
}

// Uniqueid returns the UniqueID of a signed executable.
func (c *Cli) Uniqueid(path string) (string, error) {
	data, err := c.readDataFromELF(path, oeinfoSectionName, offsetSigstruct+offsetMRENCLAVE, 32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

// Signerid returns the SignerID of a signed executable.
func (c *Cli) Signerid(path string) (string, error) {
	if filepath.Ext(path) == ".pem" {
		return c.signeridByKey(path)
	}
	return c.signeridByExecutable(path)
}
