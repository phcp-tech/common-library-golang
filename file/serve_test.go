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
	"errors"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestServeAttachment(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(path, body string) {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "test.csv"), "a,b\n1,2\n")
	write(filepath.Join(root, "sub", "报告 1.txt"), "hello")
	write(filepath.Join(parent, "secret.toml"), "token = 1")

	serve := func(name string, prepare func(w http.ResponseWriter, r *http.Request)) (*httptest.ResponseRecorder, error) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/download", nil)
		if prepare != nil {
			prepare(w, r)
		}
		return w, ServeAttachment(w, r, root, name)
	}

	t.Run("serves a file as an attachment", func(t *testing.T) {
		w, err := serve("./test.csv", nil)
		if err != nil || w.Code != http.StatusOK || w.Body.String() != "a,b\n1,2\n" {
			t.Fatalf("err %v, code %d, body %q", err, w.Code, w.Body.String())
		}
		if got := w.Header().Get("Content-Disposition"); got != "attachment; filename=test.csv" {
			t.Errorf("Content-Disposition = %q", got)
		}
	})
	t.Run("encodes a non-ASCII file name in a subdirectory", func(t *testing.T) {
		w, err := serve("sub/报告 1.txt", nil)
		if err != nil || w.Body.String() != "hello" {
			t.Fatalf("err %v, body %q", err, w.Body.String())
		}
		_, params, perr := mime.ParseMediaType(w.Header().Get("Content-Disposition"))
		if perr != nil || params["filename"] != "报告 1.txt" {
			t.Errorf("Content-Disposition = %q (%v)", w.Header().Get("Content-Disposition"), perr)
		}
	})
	t.Run("keeps the caller's Content-Type", func(t *testing.T) {
		w, _ := serve("test.csv", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		})
		if got := w.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
			t.Errorf("Content-Type = %q", got)
		}
	})
	t.Run("supports Range", func(t *testing.T) {
		w, _ := serve("test.csv", func(_ http.ResponseWriter, r *http.Request) { r.Header.Set("Range", "bytes=0-2") })
		if w.Code != http.StatusPartialContent || w.Body.String() != "a,b" {
			t.Errorf("code %d, body %q", w.Code, w.Body.String())
		}
	})
	t.Run("refuses names outside root, directories and missing files", func(t *testing.T) {
		for _, name := range []string{
			"./../secret.toml", "../secret.toml", "sub/../../secret.toml",
			filepath.Join(parent, "secret.toml"), // absolute
			"./", "sub", "missing.csv", "",
		} {
			w, err := serve(name, nil)
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("%q: err = %v, want ErrNotFound", name, err)
			}
			if w.Body.Len() != 0 || len(w.Header()) != 0 {
				t.Errorf("%q: wrote a response: %d %q %v", name, w.Code, w.Body.String(), w.Header())
			}
		}
	})
	t.Run("a missing root is not ErrNotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := ServeAttachment(w, httptest.NewRequest(http.MethodGet, "/", nil), filepath.Join(parent, "none"), "test.csv")
		if err == nil || errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v", err)
		}
	})
}
