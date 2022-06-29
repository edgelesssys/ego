// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package launch

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// UntarGzip unpacks a gzip-packed tar archive to the given path.
// Adapted from: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func UntarGzip(fs afero.Fs, r io.Reader, dst string) error {
	// Extract the runtime to a temporary directory
	compressedReader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	// Extract the tarball to the temporary directory
	tarReader := tar.NewReader(compressedReader)
	for {
		header, err := tarReader.Next()

		// When done with the archive, stop.
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header == nil {
			continue
		}

		// Construct path from the entry in the tarball
		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {

		// Directory
		case tar.TypeDir:
			if err := fs.MkdirAll(target, 0o755); err != nil {
				return err
			}

		// File
		case tar.TypeReg:
			// Tar files do not always have to have a TypeDir object in the header to specify a directory.
			// They can also be referred to implcitly by a filename with dashes.
			// Therefore, we also call MkdirAll here to create all upper directories.
			if err := fs.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}

			fsFile, err := fs.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(fsFile, tarReader); err != nil {
				return err
			}

			if err := fsFile.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}
