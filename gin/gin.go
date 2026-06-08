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
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	slogGin "github.com/samber/slog-gin"
)

// slog is used instead of the project log package so that this package has no
// initialisation dependency. slog uses the stdlib default logger when called
// standalone; when the caller invokes log.InitLog(), that function calls
// slog.SetDefault(l.Logger), so all slog calls are automatically routed to the
// project logger with no behaviour change here.

// buildOriginMatchFunc compiles a slice of wildcard patterns into an origin-matching
// function. Each pattern may contain * as a single-level wildcard (no dots), e.g.
// "http://*.localhost:5173" or "https://dev.*.example.com".
func buildOriginMatchFunc(patterns []string) func(string) bool {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		// QuoteMeta escapes all metacharacters. Replace \*\. (star-dot pair) first so
		// the subdomain is optional: "https://*.example.com" matches both "https://example.com"
		// and "https://it.example.com". Then replace any remaining \* with [^.]+ (required).
		escaped := regexp.QuoteMeta(p)
		escaped = strings.ReplaceAll(escaped, `\*\.`, `([^.]+\.)?`)
		escaped = strings.ReplaceAll(escaped, `\*`, `[^.]+`)
		regexStr := "^" + escaped + "$"
		re, err := regexp.Compile(regexStr)
		if err != nil {
			slog.Info("CORS: invalid origin pattern, skipped", "pattern", p, "error", err)
			continue
		}
		compiled = append(compiled, re)
	}
	return func(origin string) bool {
		for _, re := range compiled {
			if re.MatchString(origin) {
				return true
			}
		}
		return false
	}
}

// InitGin creates and configures a new Gin engine. It always runs in ReleaseMode
// to suppress framework debug output in all environments. It attaches structured
// slog-based request logging and a recovery middleware, and optionally configures
// CORS when corsOrigins is non-empty. Entries containing * are treated as wildcard
// patterns and matched via AllowOriginFunc (single-level wildcard * means one label,
// no dots). The caller is responsible for passing the correct origin list.
// Pass nil or an empty slice to disable CORS.
func InitGin(corsOrigins []string) *gin.Engine {
	//1. Create gin instance after SetMode
	// ReleaseMode suppresses framework route-registration logs in all environments.
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//TODO: set real IP, not available
	//router.SetTrustedProxies(nil)
	//router.TrustedPlatform = "X-Forwarded-For"

	//2. Add the slogGin middleware to all routes.
	// The middleware will log all requests attributes under a "http" group.
	// slog.Default() returns the project logger after log.InitLog() calls slog.SetDefault.
	router.Use(slogGin.New(slog.Default()))
	router.Use(gin.Recovery())

	//3. Setup cors — enabled when corsOrigins is non-empty
	if len(corsOrigins) > 0 {
		corsconfig := cors.Config{
			AllowMethods:     []string{"POST", "PUT", "DELETE", "GET", "OPTIONS", "PATCH", "HEAD"},
			AllowHeaders:     []string{"*"},
			AllowCredentials: false,          // whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates
			MaxAge:           24 * time.Hour, // how long (with second-precision) the results of a preflight request can be cached
		}

		// Split into exact origins and wildcard patterns (entries containing *)
		var exact, patterns []string
		for _, o := range corsOrigins {
			if strings.Contains(o, "*") {
				patterns = append(patterns, o)
			} else {
				exact = append(exact, o)
			}
		}
		corsconfig.AllowOrigins = exact
		if len(patterns) > 0 {
			corsconfig.AllowOriginFunc = buildOriginMatchFunc(patterns)
		}
		router.Use(cors.New(corsconfig))
	} else {
		slog.Info("CORS is not configured.")
	}

	return router
}
