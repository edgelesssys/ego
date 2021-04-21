// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package attestation

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgelesssys/ego/attestation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTLSConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	//
	// Create mock functions.
	//

	getRemoteReport := func(reportData []byte) ([]byte, error) {
		return append([]byte{2}, reportData...), nil
	}

	verifyRemoteReport := func(reportBytes []byte) (attestation.Report, error) {
		if len(reportBytes) != 33 || reportBytes[0] != 2 {
			return attestation.Report{}, errors.New("invalid remote report")
		}
		return attestation.Report{
			Data:            reportBytes[1:],
			SecurityVersion: 2,
		}, nil
	}

	failToVerifyRemoteReport := func(reportBytes []byte) (attestation.Report, error) {
		return attestation.Report{}, errors.New("invalid remote report")
	}

	verifyReport := func(report attestation.Report) error {
		if report.SecurityVersion != 2 {
			return errors.New("invalid report")
		}
		return nil
	}

	failToVerifyReport := func(report attestation.Report) error {
		return errors.New("invalid report")
	}

	//
	// Create Test Cases
	//

	tests := []struct {
		name               string
		getRemoteReport    interface{}
		verifyRemoteReport interface{}
		verifyReport       interface{}
		expectErr          bool
	}{
		{"basic", getRemoteReport, verifyRemoteReport, verifyReport, false},
		{"invalid remote report", getRemoteReport, failToVerifyRemoteReport, verifyReport, true},
		{"invalid report", getRemoteReport, verifyRemoteReport, failToVerifyReport, true},
		{"invalid remote report and report", getRemoteReport, failToVerifyRemoteReport, failToVerifyReport, true},
	}

	//
	// Run Tests.
	//

	for _, test := range tests {
		t.Logf("Subtest: %v", test.name)
		//
		// Create server.
		//

		serverConfig, err := CreateAttestationServerTLSConfig(test.getRemoteReport.(func([]byte) ([]byte, error)))
		require.NoError(err)

		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "hello")
		}))
		server.TLS = serverConfig

		//
		// Create client.
		//

		clientConfig := CreateAttestationClientTLSConfig(test.verifyRemoteReport.(func([]byte) (attestation.Report, error)), test.verifyReport.(func(attestation.Report) error))
		client := http.Client{Transport: &http.Transport{TLSClientConfig: clientConfig}}

		//
		// Test connection.
		//

		server.StartTLS()

		resp, err := client.Get(server.URL)
		if test.expectErr {
			require.Error(err)
			continue
		} else {
			require.NoError(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(err)
		assert.EqualValues("hello", body)

		server.Close()
		resp.Body.Close()
	}
}
