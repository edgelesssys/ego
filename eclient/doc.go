// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

/*
Package eclient provides functionality for Go programs that interact with enclave programs.

Use this package for programs that don't run in an enclave themselves but interact with
enclaved programs. Those non-enclaved programs are often called third parties or relying parties.

This package requires libcrypto. On Ubuntu install it with:

	sudo apt install libssl-dev

This package requires the following environment variables to be set during build:

	CGO_CFLAGS=-I/opt/ego/include
	CGO_LDFLAGS=-L/opt/ego/lib

Or if using the EGo snap:

	CGO_CFLAGS=-I/snap/ego-dev/current/opt/ego/include
	CGO_LDFLAGS=-L/snap/ego-dev/current/opt/ego/lib

For development and testing purposes, you can set the build tag `ego_mock_eclient`
instead of setting the environment variables. VerifyRemoteReport will always fail then.
*/
package eclient
