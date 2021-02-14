// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package core

import "github.com/edgelesssys/marblerun/marble/premain"

// PreMain runs before the App's actual main routine and initializes the EGo enclave.
func PreMain() error {
	// TODO
	return premain.PreMain()
}
