// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

/*
Package ecrypto provides convenience functions for cryptography inside an enclave.

# Sealing

Sealing is the process of encrypting data with a key derived from the enclave and the CPU it is running on.
Sealed data can only be decrypted by the same enclave and CPU. Use it to persist data to disk.

Use SealWithUniqueKey if the data should only be decryptable by the current enclave app version.
Use SealWithProductKey if it should also be decryptable by future versions of the enclave app.

These functions perform AES-GCM encryption. If you need something else, use the seal functions of package enclave.
*/
package ecrypto
