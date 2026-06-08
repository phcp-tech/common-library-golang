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

package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ---------------------------------------------------------------------------
// buildOriginMatchFunc
// ---------------------------------------------------------------------------

func TestBuildOriginMatchFunc(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		origin   string
		want     bool
	}{
		// *.example.com — star-dot is optional, so root domain also matches
		{"root domain matches wildcard", []string{"https://*.example.com"}, "https://example.com", true},
		{"single subdomain matches", []string{"https://*.example.com"}, "https://it.example.com", true},
		{"another subdomain matches", []string{"https://*.example.com"}, "https://edu.example.com", true},
		{"two-level subdomain no match", []string{"https://*.example.com"}, "https://dev.it.example.com", false},
		{"different domain no match", []string{"https://*.example.com"}, "https://example.org", false},
		{"http scheme no match", []string{"https://*.example.com"}, "http://example.com", false},

		// dev.*.example.com
		{"dev root domain matches", []string{"https://dev.*.example.com"}, "https://dev.example.com", true},
		{"dev subdomain matches", []string{"https://dev.*.example.com"}, "https://dev.it.example.com", true},
		{"missing dev prefix no match", []string{"https://dev.*.example.com"}, "https://it.example.com", false},

		// http://*.localhost:5173
		{"bare localhost matches", []string{"http://*.localhost:5173"}, "http://localhost:5173", true},
		{"it localhost matches", []string{"http://*.localhost:5173"}, "http://it.localhost:5173", true},
		{"edu localhost matches", []string{"http://*.localhost:5173"}, "http://edu.localhost:5173", true},
		{"wrong port no match", []string{"http://*.localhost:5173"}, "http://it.localhost:3006", false},
		{"https scheme no match", []string{"http://*.localhost:5173"}, "https://it.localhost:5173", false},

		// multiple patterns — first or second match is sufficient
		{"matches first of multiple", []string{"https://*.example.com", "https://*.example.org"}, "https://it.example.com", true},
		{"matches second of multiple", []string{"https://*.example.com", "https://*.example.org"}, "https://example.org", true},
		{"no match in multiple patterns", []string{"https://*.example.com", "https://*.example.org"}, "https://test.com", false},

		// edge cases
		{"empty patterns always false", []string{}, "https://example.com", false},
		{"empty origin always false", []string{"https://*.example.com"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := buildOriginMatchFunc(tt.patterns)
			if got := match(tt.origin); got != tt.want {
				t.Errorf("patterns=%v origin=%q: got %v, want %v", tt.patterns, tt.origin, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CORS middleware integration — verify gin-contrib/cors reflects each origin
// correctly when AllowOriginFunc is used (regression for the gin-cache
// stale-origin bug: each calling origin must receive its own origin back).
// ---------------------------------------------------------------------------

func newTestRouter(patterns []string, exact []string) *gin.Engine {
	cfg := cors.Config{
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: false,
		MaxAge:           24 * time.Hour,
	}
	cfg.AllowOrigins = exact
	if len(patterns) > 0 {
		cfg.AllowOriginFunc = buildOriginMatchFunc(patterns)
	}
	r := gin.New()
	r.Use(cors.New(cfg))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

// newCORSRequest creates a test request with an explicit Host header set to
// "testserver" so that gin-contrib/cors does not mistake the tested origin for
// the server's own host (same-origin check: origin == "https://"+host).
func newCORSRequest(method, path, origin string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.Host = "testserver"
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	return req
}

func assertCORSOrigin(t *testing.T, router *gin.Engine, origin, wantOrigin string) {
	t.Helper()
	w := httptest.NewRecorder()
	req := newCORSRequest(http.MethodGet, "/ping", origin)
	router.ServeHTTP(w, req)
	got := w.Header().Get("Access-Control-Allow-Origin")
	if got != wantOrigin {
		t.Errorf("Origin: %q → Access-Control-Allow-Origin = %q, want %q", origin, got, wantOrigin)
	}
}

func TestCORSMiddleware_ExactOrigin(t *testing.T) {
	router := newTestRouter(nil, []string{"https://example.com"})
	assertCORSOrigin(t, router, "https://example.com", "https://example.com")
	assertCORSOrigin(t, router, "https://test.com", "")
}

func TestCORSMiddleware_WildcardPatternReflectsRequestOrigin(t *testing.T) {
	// Core regression test: three different subdomain origins must each receive
	// their own origin in Access-Control-Allow-Origin, not the first caller's.
	router := newTestRouter([]string{"https://*.example.com"}, nil)

	origins := []string{
		"https://edu.example.com",
		"https://it.example.com",
		"https://lit.example.com",
		"https://example.com",
	}
	for _, origin := range origins {
		t.Run(origin, func(t *testing.T) {
			assertCORSOrigin(t, router, origin, origin)
		})
	}
}

func TestCORSMiddleware_WildcardPatternBlocksUnknownOrigin(t *testing.T) {
	router := newTestRouter([]string{"https://*.example.com"}, nil)
	assertCORSOrigin(t, router, "https://test.com", "")
}

func TestCORSMiddleware_MixedExactAndWildcard(t *testing.T) {
	// Exact origins and wildcard patterns can coexist in config.
	router := newTestRouter(
		[]string{"https://*.example.org"},
		[]string{"https://example.com"},
	)
	assertCORSOrigin(t, router, "https://example.com", "https://example.com")       // exact
	assertCORSOrigin(t, router, "https://example.org", "https://example.org")       // wildcard root
	assertCORSOrigin(t, router, "https://it.example.org", "https://it.example.org") // wildcard subdomain
	assertCORSOrigin(t, router, "https://test.com", "")                             // blocked
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	router := newTestRouter([]string{"https://*.example.com"}, nil)

	w := httptest.NewRecorder()
	req := newCORSRequest(http.MethodOptions, "/ping", "https://it.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://it.example.com" {
		t.Errorf("preflight: unexpected Access-Control-Allow-Origin: %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("preflight: unexpected status %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// InitGin — engine creation and CORS integration
// ---------------------------------------------------------------------------

func TestInitGin_ReturnsNonNilEngine(t *testing.T) {
	router := InitGin(nil)
	if router == nil {
		t.Fatal("InitGin returned nil engine")
	}
}

// TestInitGin_NoOrigins_NoCORSHeader verifies that nil and empty-slice origins
// both disable CORS — len(nil) == len([]string{}) == 0 in Go, same code path.
func TestInitGin_NoOrigins_NoCORSHeader(t *testing.T) {
	cases := []struct {
		name    string
		origins []string
	}{
		{"nil", nil},
		{"empty slice", []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			router := InitGin(c.origins)
			router.GET("/ping", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			router.ServeHTTP(w, newCORSRequest(http.MethodGet, "/ping", "https://example.com"))

			if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
				t.Errorf("expected no CORS header, got %q", got)
			}
		})
	}
}

func TestInitGin_ExactOrigin_CORSEnabled(t *testing.T) {
	router := InitGin([]string{"https://example.com"})
	router.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	router.ServeHTTP(w, newCORSRequest(http.MethodGet, "/ping", "https://example.com"))

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("Access-Control-Allow-Origin: want %q, got %q", "https://example.com", got)
	}
}

func TestInitGin_WildcardOrigin_CORSEnabled(t *testing.T) {
	router := InitGin([]string{"https://*.example.com"})
	router.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	origins := []string{"https://example.com", "https://api.example.com", "https://dev.example.com"}
	for _, origin := range origins {
		t.Run(origin, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, newCORSRequest(http.MethodGet, "/ping", origin))
			if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
				t.Errorf("Access-Control-Allow-Origin: want %q, got %q", origin, got)
			}
		})
	}
}

func TestInitGin_BlockedOrigin_NoCORSHeader(t *testing.T) {
	router := InitGin([]string{"https://example.com"})
	router.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	router.ServeHTTP(w, newCORSRequest(http.MethodGet, "/ping", "https://test.com"))

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("blocked origin: expected no CORS header, got %q", got)
	}
}
