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

package download

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func newRoot(t *testing.T) (parent, root string) {
	t.Helper()
	parent = t.TempDir()
	root = filepath.Join(parent, "root")
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "test.csv"), []byte("a,b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "secret.toml"), []byte("token = 1"), 0o644); err != nil {
		t.Fatal(err)
	}
	return parent, root
}

func get(r http.Handler, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
	return w
}

func code(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q is not JSON: %v", w.Body.String(), err)
	}
	return body.Code
}

func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, root := newRoot(t)
	r := gin.New()
	r.GET("/download", Handler(root))
	r.GET("/csv", func(c *gin.Context) {
		c.Header("Content-Type", "text/csv; charset=utf-8")
		Handler(root)(c)
	})

	w := get(r, "/download?file=./test.csv")
	if w.Code != http.StatusOK || w.Body.String() != "a,b\n" {
		t.Fatalf("download: %d %q", w.Code, w.Body.String())
	}
	if w.Header().Get("Content-Disposition") != "attachment; filename=test.csv" ||
		w.Header().Get("Access-Control-Expose-Headers") != "Content-Disposition" {
		t.Errorf("headers = %v", w.Header())
	}
	if w = get(r, "/csv?file=./test.csv"); w.Header().Get("Content-Type") != "text/csv; charset=utf-8" {
		t.Errorf("caller's Content-Type = %q", w.Header().Get("Content-Type"))
	}

	if w = get(r, "/download?file=test.csv"); w.Code != http.StatusBadRequest || code(t, w) != http.StatusBadRequest {
		t.Errorf("no ./ prefix: %d %q", w.Code, w.Body.String())
	}
	for _, name := range []string{"./missing.csv", "./", "./sub", "./../secret.toml", "./sub/../../secret.toml"} {
		w = get(r, "/download?file="+url.QueryEscape(name))
		if w.Code != http.StatusNotFound || code(t, w) != http.StatusNotFound || w.Header().Get("Content-Disposition") != "" {
			t.Errorf("%s: %d %q %v", name, w.Code, w.Body.String(), w.Header())
		}
	}

	bad := gin.New()
	bad.GET("/download", Handler(filepath.Join(root, "none")))
	if w = get(bad, "/download?file=./test.csv"); w.Code != http.StatusInternalServerError || code(t, w) != http.StatusInternalServerError {
		t.Errorf("missing root: %d %q", w.Code, w.Body.String())
	}
}

func TestMount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, root := newRoot(t)
	deny := func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) }
	allow := func(c *gin.Context) { c.Next() }

	r := gin.New()
	if Mount(r, "/api/v1", root, deny) != r {
		t.Fatal("Mount returned another engine")
	}
	if w := get(r, "/api/v1/admin/files/download?file=./test.csv"); w.Code != http.StatusUnauthorized {
		t.Errorf("guarded route: %d", w.Code)
	}

	r = gin.New()
	Mount(r, "/api/v1", root, allow)
	if w := get(r, "/api/v1/admin/files/download?file=./test.csv"); w.Code != http.StatusOK || w.Body.String() != "a,b\n" {
		t.Errorf("allowed route: %d %q", w.Code, w.Body.String())
	}

	defer func() {
		if recover() == nil {
			t.Error("Mount without handlers did not panic")
		}
	}()
	Mount(gin.New(), "/api/v1", root)
}

func TestExeDir(t *testing.T) {
	dir, err := ExeDir()
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("ExeDir = %q: %v", dir, err)
	}
}
