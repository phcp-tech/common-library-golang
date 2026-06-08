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

// Package pprof exposes Go runtime profiling endpoints on a Gin engine via
// github.com/gin-contrib/pprof. Import this package only in services that
// require profiling — importing it registers pprof handlers as a side effect.
package pprof

import (
	contribpprof "github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

// Mount registers pprof profiling endpoints on the provided Gin engine.
// Each handler is a native Gin route, so all Gin middleware (auth, rate-limit,
// etc.) applies cleanly.
//
// Two prefix groups are registered:
//   - /debug/pprof/*        — standard Go pprof path for direct access
//   - <path>/admin/pprof/*  — API-gateway-friendly alias; <path> is the gateway
//     prefix forwarded to this service (e.g. "/api/v1")
func Mount(router *gin.Engine, path string) *gin.Engine {
	// URL1: standard pprof endpoint for direct access
	contribpprof.Register(router)

	// URL2: API-gateway-friendly alias so that <path>/admin/pprof/* also works
	contribpprof.Register(router, path+"/admin/pprof")

	return router
}
