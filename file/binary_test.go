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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsBinary(t *testing.T) {
	zh := strings.Repeat("知识", 200) // 1200 bytes; byte 512 cuts a character
	cases := map[string]struct {
		body string
		want bool
	}{
		"empty":                           {"", false},
		"ASCII text":                      {"a,b\r\n1,2\tx\n", false},
		"Chinese text over 512 bytes":     {zh, false},
		"emoji cut at byte 512":           {"a" + strings.Repeat("😀", 200), false},
		"NUL byte":                        {"abc\x00def", true},
		"NUL byte after byte 512":         {strings.Repeat("a", sniffSize) + "\x00", false},
		"PNG header":                      {"\x89PNG\r\n\x1a\n" + strings.Repeat("\xff\xfe", 100), true},
		"Latin-1 text, few high bytes":    {strings.Repeat("a", 100) + "\xe9", false},
		"invalid UTF-8, exactly 5%":       {strings.Repeat("a", 95) + strings.Repeat("\xe9", 5), false},
		"invalid UTF-8, over 5%":          {strings.Repeat("a", 94) + strings.Repeat("\xe9", 6), true},
		"invalid UTF-8, many high bytes":  {strings.Repeat("\xc3\x28", 50), true},
		"control bytes count as non-text": {strings.Repeat("a", 94) + "\xe9" + strings.Repeat("\x01", 5), true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			r := bytes.NewReader([]byte(c.body))
			r.Seek(int64(len(c.body)/2), io.SeekStart) // IsBinary must not depend on the current position
			got, err := IsBinary(r)
			if err != nil || got != c.want {
				t.Errorf("IsBinary = %v, %v; want %v", got, err, c.want)
			}
			if pos, _ := r.Seek(0, io.SeekCurrent); pos != 0 {
				t.Errorf("position after IsBinary = %d, want 0", pos)
			}
		})
	}
}

func TestIsBinary_FileCanBeReadAgain(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.csv")
	body := strings.Repeat("a,b\n", 300)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if got, err := IsBinary(f); err != nil || got {
		t.Fatalf("IsBinary = %v, %v; want false", got, err)
	}
	if b, err := io.ReadAll(f); err != nil || string(b) != body {
		t.Errorf("read after IsBinary: %d bytes, %v; want %d bytes", len(b), err, len(body))
	}
}

// faultyReader fails Read or Seek on demand.
type faultyReader struct {
	io.ReadSeeker
	readErr, seekErr error
}

func (f faultyReader) Read(p []byte) (int, error) {
	if f.readErr != nil {
		return 0, f.readErr
	}
	return f.ReadSeeker.Read(p)
}

func (f faultyReader) Seek(offset int64, whence int) (int64, error) {
	if f.seekErr != nil {
		return 0, f.seekErr
	}
	return f.ReadSeeker.Seek(offset, whence)
}

func TestIsBinary_Errors(t *testing.T) {
	errDisk := errors.New("disk error")
	for name, r := range map[string]faultyReader{
		"read fails": {ReadSeeker: strings.NewReader("abc"), readErr: errDisk},
		"seek fails": {ReadSeeker: strings.NewReader("abc"), seekErr: errDisk},
	} {
		if _, err := IsBinary(r); !errors.Is(err, errDisk) {
			t.Errorf("%s: err = %v, want %v", name, err, errDisk)
		}
	}
}

func TestTrimPartialRune(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"empty":                     {"", ""},
		"ASCII":                     {"abc", "abc"},
		"full 3-byte character":     {"a知", "a知"},
		"cut 3-byte character, 2/3": {"a\xe7\x9f", "a"},
		"cut 3-byte character, 1/3": {"a\xe7", "a"},
		"cut 4-byte character, 3/4": {"a\xf0\x9f\x98", "a"},
		"full 4-byte character":     {"a😀", "a😀"},
		"continuation bytes only":   {"\x80\x80\x80\x80\x80", "\x80\x80\x80\x80\x80"},
	}
	for name, c := range cases {
		if got := string(trimPartialRune([]byte(c.in))); got != c.want {
			t.Errorf("%s: trimPartialRune(%q) = %q, want %q", name, c.in, got, c.want)
		}
	}
}
