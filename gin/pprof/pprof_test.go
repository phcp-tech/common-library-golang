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

package pprof_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	libpprof "github.com/phcp-tech/common-library-golang/gin/pprof"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newEngine() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	return r
}

// -----------------------------------------------------------------------
// Mount — return value
// -----------------------------------------------------------------------

func TestMount_ReturnsSameEngine(t *testing.T) {
	r := newEngine()
	got := libpprof.Mount(r, "/api/v1")
	if got != r {
		t.Error("Mount should return the same *gin.Engine it received")
	}
}

// -----------------------------------------------------------------------
// Mount — URL1: standard /debug/pprof/* endpoint
// -----------------------------------------------------------------------

func TestMount_StandardEndpoint_Index(t *testing.T) {
	r := newEngine()
	libpprof.Mount(r, "/api/v1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/debug/pprof/ status: want 200, got %d", w.Code)
	}
}

func TestMount_StandardEndpoint_Profiles(t *testing.T) {
	r := newEngine()
	libpprof.Mount(r, "/api/v1")

	endpoints := []string{
		"/debug/pprof/cmdline",
		"/debug/pprof/symbol",
		"/debug/pprof/heap",
		"/debug/pprof/goroutine",
	}
	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, ep, nil)
			r.ServeHTTP(w, req)
			if w.Code == http.StatusNotFound {
				t.Errorf("%s: got 404, pprof endpoint not registered", ep)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Mount — URL2: API-gateway alias <path>/admin/pprof/*
// -----------------------------------------------------------------------

func TestMount_GatewayAlias_Index(t *testing.T) {
	r := newEngine()
	libpprof.Mount(r, "/api/v1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/pprof/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/api/v1/admin/pprof/ status: want 200, got %d", w.Code)
	}
}

func TestMount_GatewayAlias_Profiles(t *testing.T) {
	r := newEngine()
	libpprof.Mount(r, "/api/v1")

	endpoints := []string{
		"/api/v1/admin/pprof/cmdline",
		"/api/v1/admin/pprof/heap",
		"/api/v1/admin/pprof/goroutine",
	}
	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, ep, nil)
			r.ServeHTTP(w, req)
			if w.Code == http.StatusNotFound {
				t.Errorf("%s: got 404, pprof alias endpoint not registered", ep)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Mount — different gateway path prefixes
// -----------------------------------------------------------------------

func TestMount_CustomGatewayPath(t *testing.T) {
	cases := []struct {
		path     string
		expected string
	}{
		{"/v2", "/v2/admin/pprof/"},
		{"/service/orders", "/service/orders/admin/pprof/"},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			r := newEngine()
			libpprof.Mount(r, c.path)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, c.expected, nil)
			r.ServeHTTP(w, req)
			if w.Code == http.StatusNotFound {
				t.Errorf("%s: got 404, expected pprof alias to be registered", c.expected)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Mount — response content sanity check
// -----------------------------------------------------------------------

func TestMount_IndexContainsPprofLinks(t *testing.T) {
	r := newEngine()
	libpprof.Mount(r, "/api/v1")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "goroutine") {
		t.Error("/debug/pprof/ response body should contain 'goroutine'")
	}
	if !strings.Contains(body, "heap") {
		t.Error("/debug/pprof/ response body should contain 'heap'")
	}
}
