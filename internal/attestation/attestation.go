// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package attestation

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"math/big"
	"time"

	"github.com/edgelesssys/ego/attestation/tcbstatus"
)

// Report is a parsed enclave report.
type Report struct {
	Data            []byte           // The report data that has been included in the report.
	SecurityVersion uint             // Security version of the enclave. For SGX enclaves, this is the ISVSVN value.
	Debug           bool             // If true, the report is for a debug enclave.
	UniqueID        []byte           // The unique ID for the enclave. For SGX enclaves, this is the MRENCLAVE value.
	SignerID        []byte           // The signer ID for the enclave. For SGX enclaves, this is the MRSIGNER value.
	ProductID       []byte           // The Product ID for the enclave. For SGX enclaves, this is the ISVPRODID value.
	TCBStatus       tcbstatus.Status // The status of the enclave's TCB level.
}

// https://github.com/openenclave/openenclave/blob/master/include/openenclave/internal/report.h
var oidOeNewQuote = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 105, 1}

func hashPublicKey(pub interface{}) ([]byte, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	result := sha256.Sum256(pubBytes)
	return result[:], nil
}

// CreateAttestationCertificate creates an X.509 certificate with an embedded report from getRemoteReport.
func CreateAttestationCertificate(getRemoteReport func([]byte) ([]byte, error), template, parent *x509.Certificate, pub, priv interface{}) ([]byte, error) {
	// get report for the public key
	hash, err := hashPublicKey(pub)
	if err != nil {
		return nil, err
	}
	report, err := getRemoteReport(hash)
	if err != nil {
		return nil, err
	}

	template.ExtraExtensions = append(template.ExtraExtensions, pkix.Extension{Id: oidOeNewQuote, Value: report})

	return x509.CreateCertificate(rand.Reader, template, parent, pub, priv)
}

// CreateAttestationServerTLSConfig creates a tls.Config object with a self-signed certificate and an embedded report.
func CreateAttestationServerTLSConfig(getRemoteReport func([]byte) ([]byte, error)) (*tls.Config, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{CommonName: "EGo"},
		NotAfter:     time.Now().AddDate(1, 0, 0),
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	cert, err := CreateAttestationCertificate(getRemoteReport, template, template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{cert},
				PrivateKey:  priv,
			},
		},
	}, nil
}

// CreateAttestationClientTLSConfig creates a tls.Config object that verifies a certificate with embedded report.
func CreateAttestationClientTLSConfig(verifyRemoteReport func([]byte) (Report, error), opts Options, verifyReport func(Report) error) *tls.Config {
	verify := func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		// parse certificate
		if len(rawCerts) <= 0 {
			return errors.New("rawCerts is empty")
		}
		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return err
		}

		// verify self-signed certificate
		roots := x509.NewCertPool()
		roots.AddCert(cert)
		_, err = cert.Verify(x509.VerifyOptions{Roots: roots})
		if err != nil {
			return err
		}

		hash, err := hashPublicKey(cert.PublicKey)
		if err != nil {
			return err
		}

		// verify embedded report
		for _, ex := range cert.Extensions {
			if ex.Id.Equal(oidOeNewQuote) {
				report, err := verifyRemoteReport(ex.Value)
				if err != nil && err != opts.IgnoreErr {
					return err
				}
				if !bytes.Equal(report.Data[:len(hash)], hash) {
					return errors.New("certificate hash does not match report data")
				}
				return verifyReport(report)
			}
		}

		return errors.New("certificate does not contain attestation report")
	}

	return &tls.Config{VerifyPeerCertificate: verify, InsecureSkipVerify: true}
}

// Options are attestation options.
type Options struct {
	IgnoreErr error
}
