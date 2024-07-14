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
	"fmt"
	"io"
	"os"
	"strings"
)

const oeinfoSectionName = ".oeinfo"

// ErrErrUnsupportedImportEClient is returned when an EGo binary uses the eclient package instead of the enclave package.
var ErrUnsupportedImportEClient = errors.New("unsupported import: github.com/edgelesssys/ego/eclient")

// ErrLargeHeapWithSmallHeapSize is returned when a binary is built with ego_largehap, but the heap size is set to less than 512MB.
var ErrLargeHeapWithSmallHeapSize = errors.New("large heap is enabled, but heapSize is too small")

// ErrNoLargeHeapWithLargeHeapSize is returned when a binary is built without ego_largeheap, but the heap size is set to more than 16GB.
var ErrNoLargeHeapWithLargeHeapSize = errors.New("this heapSize requires large heap mode")

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

	oeInfo := elfFile.Section(oeinfoSectionName)
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

func (c *Cli) getSymbolsFromELF(path string) ([]elf.Symbol, error) {
	file, err := c.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	elfFile, err := elf.NewFile(file)
	if err != nil {
		return nil, err
	}

	return elfFile.Symbols()
}

func (c *Cli) readDataFromELF(path string, section string, offset int, size int) ([]byte, error) {
	file, err := c.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	elfFile, err := elf.NewFile(file)
	if err != nil {
		return nil, err
	}

	sec := elfFile.Section(section)
	if sec == nil {
		return nil, errors.New("section not found")
	}

	data := make([]byte, size)
	if _, err := sec.ReadAt(data, int64(offset)); err != nil {
		return nil, err
	}

	return data, nil
}

// checkUnsupportedImports checks whether the to-be-signed or to-be-executed binary uses Go imports which are not supported.
func (c *Cli) checkUnsupportedImports(path string) error {
	symbols, err := c.getSymbolsFromELF(path)
	if err == nil {
		return checkUnsupportedImports(symbols)
	}
	if errors.Is(err, elf.ErrNoSymbols) {
		return nil
	}
	return fmt.Errorf("getting symbols: %w", err)
}

func checkUnsupportedImports(symbols []elf.Symbol) error {
	// Iterate through all symbols and find whether it matches a known unsupported one
	for _, symbol := range symbols {
		if strings.Contains(symbol.Name, "github.com/edgelesssys/ego/eclient") {
			return ErrUnsupportedImportEClient
		}
	}
	return nil
}

// checkHeapMode checks whether the heapSize is compatible with the binary.
// If it is built with the ego_largeheap build tag, heapSize must be >= 512.
// If it is built without this tag, heapSize must be <= 16384.
// (If 512 <= heapSize <= 16384, both modes work.)
func checkHeapMode(symbols []elf.Symbol, heapSize int) error {
	for _, symbol := range symbols {
		if symbol.Name == "runtime.arenaBaseOffset" {
			// if this symbol is found, the binary wasn't built with ego_largeheap
			if heapSize > 16384 {
				return ErrNoLargeHeapWithLargeHeapSize
			}
			return nil
		}
	}
	// if the symbol isn't found, the binary was built with ego_largeheap
	if heapSize < 512 {
		return ErrLargeHeapWithSmallHeapSize
	}
	return nil
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
