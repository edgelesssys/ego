// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Reserve TLS space for payload. See comment about reserved_tls lib in
// CMakeLists.
__thread char ert_reserved_tls[11264];
