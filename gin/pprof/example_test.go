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

// package pprof_test demonstrates the public API from a caller's perspective.
package pprof_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	libgin "github.com/phcp-tech/common-library-golang/gin"
	libpprof "github.com/phcp-tech/common-library-golang/gin/pprof"
)

// ExampleMount shows how to register pprof endpoints on a Gin engine.
// Mount registers two sets of routes:
//   - /debug/pprof/*            — standard Go pprof path
//   - <path>/admin/pprof/*      — API-gateway-friendly alias
//
// Import this package only in services that require profiling.
func ExampleMount() {
	gin.SetMode(gin.TestMode)

	router := libgin.InitGin(nil)
	libpprof.Mount(router, "/api/v1")

	// Standard pprof index is reachable at /debug/pprof/.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	router.ServeHTTP(w, req)
	fmt.Println(w.Code)

	// API-gateway alias is also reachable.
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/pprof/", nil)
	router.ServeHTTP(w2, req2)
	fmt.Println(w2.Code)
	// Output:
	// 200
	// 200
}
