// Copyright (c) Edgeless Systems GmbH.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package launch

import (
	"bytes"
	"io"
	"log/slog"
	"regexp"
)

var rxLogTag = regexp.MustCompile(`^\[(erthost|ego)\] `)

// jsonifier is a writer that rewrites ert/ego log messages to JSON.
type jsonifier struct {
	out io.Writer
	log *slog.Logger
	buf []byte
}

func newJsonifier(out io.Writer) *jsonifier {
	return &jsonifier{
		out: out,
		log: slog.New(slog.NewJSONHandler(out, nil)),
	}
}

func (j *jsonifier) Write(p []byte) (int, error) {
	buf := append(j.buf, p...)

	// write out all complete lines in buf
	for {
		idx := bytes.IndexByte(buf, '\n')
		if idx == -1 {
			break
		}

		// use JSON logger if line begins with known tag
		if rxLogTag.Match(buf[:idx]) {
			j.log.Info(string(buf[:idx]))
		} else {
			_, _ = j.out.Write(buf[:idx+1])
		}

		buf = buf[idx+1:]
	}

	j.buf = buf
	return len(p), nil
}
