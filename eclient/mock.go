// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build ego_mock_eclient

package eclient

import (
	"errors"

	"github.com/edgelesssys/ego/internal/attestation"
)

func verifyRemoteReport([]byte) (attestation.Report, error) {
	return attestation.Report{}, errors.New("built with ego_mock_eclient tag, no attestation support available")
}
