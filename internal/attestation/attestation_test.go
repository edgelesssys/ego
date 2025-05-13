// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package attestation

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTLSConfig(t *testing.T) {
	//
	// Create mock functions.
	//

	getRemoteReport := func(reportData []byte) ([]byte, error) {
		return append([]byte{2}, reportData...), nil
	}

	verifyRemoteReport := func(reportBytes []byte) (Report, error) {
		if len(reportBytes) != 33 || reportBytes[0] != 2 {
			return Report{}, errors.New("invalid remote report")
		}
		return Report{
			Data:            reportBytes[1:],
			SecurityVersion: 2,
		}, nil
	}

	failToVerifyRemoteReportErr := errors.New("invalid remote report")
	failToVerifyRemoteReport := func(reportBytes []byte) (Report, error) {
		return Report{
			Data:            reportBytes[1:],
			SecurityVersion: 2,
		}, failToVerifyRemoteReportErr
	}

	verifyReport := func(report Report) error {
		if report.SecurityVersion != 2 {
			return errors.New("invalid report")
		}
		return nil
	}

	failToVerifyReport := func(report Report) error {
		return errors.New("invalid report")
	}

	//
	// Create Test Cases
	//

	testCases := map[string]struct {
		getRemoteReport    func([]byte) ([]byte, error)
		verifyRemoteReport func([]byte) (Report, error)
		opts               Options
		verifyReport       func(Report) error
		wantErr            bool
	}{
		"basic": {
			getRemoteReport:    getRemoteReport,
			verifyRemoteReport: verifyRemoteReport,
			verifyReport:       verifyReport,
		},
		"invalid remote report": {
			getRemoteReport:    getRemoteReport,
			verifyRemoteReport: failToVerifyRemoteReport,
			verifyReport:       verifyReport,
			wantErr:            true,
		},
		"invalid report": {
			getRemoteReport:    getRemoteReport,
			verifyRemoteReport: verifyRemoteReport,
			verifyReport:       failToVerifyReport,
			wantErr:            true,
		},
		"invalid remote report and report": {
			getRemoteReport: getRemoteReport, verifyRemoteReport: failToVerifyRemoteReport,
			verifyReport: failToVerifyReport,
			wantErr:      true,
		},
		"ignore remote report error": {
			getRemoteReport:    getRemoteReport,
			verifyRemoteReport: failToVerifyRemoteReport,
			opts:               Options{IgnoreErr: failToVerifyRemoteReportErr},
			verifyReport:       verifyReport,
		},
		"ignore other remote report error": {
			getRemoteReport:    getRemoteReport,
			verifyRemoteReport: failToVerifyRemoteReport,
			opts:               Options{IgnoreErr: assert.AnError},
			verifyReport:       verifyReport,
			wantErr:            true,
		},
	}

	//
	// Run Tests.
	//

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			//
			// Create server.
			//

			serverConfig, err := CreateAttestationServerTLSConfig(tc.getRemoteReport)
			require.NoError(err)

			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, "hello")
			}))
			server.TLS = serverConfig

			//
			// Create client.
			//

			clientConfig := CreateAttestationClientTLSConfig(tc.verifyRemoteReport, tc.opts, tc.verifyReport)
			client := http.Client{Transport: &http.Transport{TLSClientConfig: clientConfig}}

			//
			// Test connection.
			//

			server.StartTLS()
			defer server.Close()

			resp, err := client.Get(server.URL)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(err)
			assert.EqualValues("hello", body)
		})
	}
}
