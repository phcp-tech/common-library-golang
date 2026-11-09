// Copyright(C) 2019-2026 PHCP Technologies. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

// sniffSize is how many leading bytes IsBinary reads.
const sniffSize = 512

// IsBinary reports whether f looks like a binary file, judging by its first
// 512 bytes: a NUL byte means binary; valid UTF-8 means text; otherwise more
// than 5% of bytes outside printable ASCII (tab, LF and CR count as text)
// means binary. A multi-byte UTF-8 character cut at the 512-byte boundary does
// not make text look binary. An empty file is text.
//
// f is read from its start and left at its start, so a multipart.File or an
// *os.File can be read again afterwards.
func IsBinary(f io.ReadSeeker) (bool, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return false, err
	}
	buf := make([]byte, sniffSize)
	n, err := io.ReadFull(f, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return false, err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return false, err
	}
	buf = buf[:n]
	if n == sniffSize {
		buf = trimPartialRune(buf)
	}

	if bytes.IndexByte(buf, 0) >= 0 {
		return true, nil
	}
	if utf8.Valid(buf) {
		return false, nil
	}
	nonText := 0
	for _, b := range buf {
		if (b < 32 && b != '\t' && b != '\n' && b != '\r') || b > 127 {
			nonText++
		}
	}
	return nonText > len(buf)*5/100, nil
}

// trimPartialRune drops an incomplete UTF-8 character at the end of buf.
func trimPartialRune(buf []byte) []byte {
	for i := len(buf) - 1; i >= 0 && i >= len(buf)-utf8.UTFMax; i-- {
		if utf8.RuneStart(buf[i]) {
			if !utf8.FullRune(buf[i:]) {
				return buf[:i]
			}
			break
		}
	}
	return buf
}
