// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"debug/elf"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

func (c *Cli) embedConfigAsPayload(path string, jsonData []byte) error {
	// Load ELF executable
	f, err := c.fs.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	// Check if a payload already exists
	payloadSize, payloadOffset, oeInfoOffset, err := getPayloadInformation(f)
	if err != nil {
		return err
	}

	// If a payload already exists, truncate the file to remove it
	if payloadSize > 0 {
		fileStat, err := f.Stat()
		if err != nil {
			return err
		}

		// Check if payload is at expected location
		expectedPayloadOffset := fileStat.Size() - int64(payloadSize)
		if expectedPayloadOffset != payloadOffset {
			return errors.New("expected payload location does not match real payload location, cannot safely truncate old payload")
		}

		err = f.Truncate(payloadOffset)
		if err != nil {
			return err
		}
	} else if (payloadSize == 0) != (payloadOffset == 0) {
		return errors.New("payload information in header is inconsistent, cannot continue")
	}

	// Get current file size to determine offset
	fileStat, err := f.Stat()
	if err != nil {
		return err
	}
	filesize := fileStat.Size()

	// Write payload offset and size to .oeinfo header
	if err := writeUint64At(f, uint64(filesize), oeInfoOffset+2048); err != nil {
		return err
	}
	if err := writeUint64At(f, uint64(len(jsonData)), oeInfoOffset+2056); err != nil {
		return err
	}

	// And finally, append the payload to the file
	n, err := f.WriteAt(jsonData, filesize)
	if err != nil {
		return err
	} else if n != len(jsonData) {
		return errors.New("failed to embed enclave.json as metadata")
	}

	return nil
}

func getPayloadInformation(f io.ReaderAt) (uint64, int64, int64, error) {
	// .oeinfo + 2056 contains the size of an embedded Edgeless RT data payload.
	// If it is > 0, a payload already exists.

	elfFile, err := elf.NewFile(f)
	if err != nil {
		return 0, 0, 0, err
	}

	oeInfo := elfFile.Section(".oeinfo")
	if oeInfo == nil {
		return 0, 0, 0, ErrNoOEInfo
	}

	payloadOffset, err := readUint64At(oeInfo, 2048)
	if err != nil {
		return 0, 0, 0, err
	}
	payloadSize, err := readUint64At(oeInfo, 2056)
	if err != nil {
		return 0, 0, 0, err
	}

	return payloadSize, int64(payloadOffset), int64(oeInfo.Offset), nil
}

func writeUint64At(w io.WriterAt, x uint64, off int64) error {
	xByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(xByte, x)

	n, err := w.WriteAt(xByte, off)
	if err != nil {
		return err
	} else if n != 8 {
		return errors.New("did not write expected number of bytes")
	}

	return nil
}

func readUint64At(r io.ReaderAt, off int64) (uint64, error) {
	xByte := make([]byte, 8)

	n, err := r.ReadAt(xByte, off)
	if err != nil {
		return 0, err
	} else if n != 8 {
		return 0, errors.New("did not read expected number of bytes")
	}

	return binary.LittleEndian.Uint64(xByte), nil
}
