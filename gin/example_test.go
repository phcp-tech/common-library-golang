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

// package gin_test demonstrates the public API from a caller's perspective.
package gin_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	libgin "github.com/phcp-tech/common-library-golang/gin"
	slogGin "github.com/samber/slog-gin"
)

// ExampleInitGin shows how to create a Gin engine with no CORS.
// Pass nil or an empty slice when all cross-origin requests should be blocked.
func ExampleInitGin() {
	router := libgin.InitGin(nil)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	fmt.Println(router != nil)
	// Output:
	// true
}

// ExampleInitGin_cors shows how to enable CORS for a list of allowed origins.
// Exact origins and wildcard patterns (entries containing *) can be mixed.
// A wildcard * matches exactly one subdomain label (no dots).
func ExampleInitGin_cors() {
	router := libgin.InitGin([]string{
		"https://app.example.com", // exact origin
		"https://*.example.com",   // wildcard: root + any single subdomain
	})
	router.GET("/api/data", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Allowed origin receives its own origin in the response header.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Origin", "https://app.example.com")
	router.ServeHTTP(w, req)
	fmt.Println(w.Header().Get("Access-Control-Allow-Origin"))

	// Blocked origin receives no CORS header.
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req2.Header.Set("Origin", "https://test.com")
	router.ServeHTTP(w2, req2)
	fmt.Println(w2.Header().Get("Access-Control-Allow-Origin"))
	// Output:
	// https://app.example.com
	//
}

// ExampleInitGin_filters shows how to silence request logging for specific
// endpoints - e.g. a noisy health-check hit repeatedly by a load balancer, or
// a metrics-scrape path. filters is a trailing variadic (InitGin(corsOrigins
// []string, filters ...slogGin.Filter)), so filtering several endpoints
// needs no new plumbing - either pass multiple slogGin.Filter values, or a
// single IgnorePath(...) call listing several paths (IgnorePath itself is
// variadic too); the two styles below are equivalent. A request is logged
// unless ANY filter returns false for it - filters combine with OR, not AND.
func ExampleInitGin_filters() {
	// Route logging to a buffer instead of stderr so this example can
	// assert on it. Real callers never do this - see log.InitLog(), which
	// InitGin's own doc comment mentions is what redirects slog.Default().
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))

	router := libgin.InitGin(nil,
		slogGin.IgnorePath("/healthz", "/readyz"), // one call, multiple paths
		slogGin.IgnorePathPrefix("/metrics"),      // a second, distinct filter
	)
	for _, path := range []string{"/healthz", "/readyz", "/metrics/cpu", "/api/data"} {
		router.GET(path, func(c *gin.Context) { c.Status(http.StatusOK) })
	}

	for _, path := range []string{"/healthz", "/readyz", "/metrics/cpu", "/api/data"} {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	fmt.Println(strings.Contains(buf.String(), "/healthz"))
	fmt.Println(strings.Contains(buf.String(), "/readyz"))
	fmt.Println(strings.Contains(buf.String(), "/metrics/cpu"))
	fmt.Println(strings.Contains(buf.String(), "/api/data"))
	// Output:
	// false
	// false
	// false
	// true
}
