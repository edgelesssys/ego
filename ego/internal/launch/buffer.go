// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package launch

// cappedBuffer is a buffer that only takes up to 1000 bytes in total.
type cappedBuffer []byte

// Write writes until the CappedBuffer is full and silently discards the remainder.
func (b *cappedBuffer) Write(p []byte) (int, error) {
	if len(*b) < 1000 {
		*b = append(*b, p...)
	}
	return len(p), nil
}
