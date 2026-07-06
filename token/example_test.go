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

// package token_test demonstrates the public API from a caller's perspective.
// The JWT secrets used here are initialised by TestMain in token_test.go.
package token_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phcp-tech/common-library-golang/token"
)

// ExampleInitToken shows how to initialise the token package once at application
// startup. The secrets and issuer are typically read from env.Env() after
// env.InitEnv() has been called.
func ExampleInitToken() {
	token.InitToken(
		"myapp",             // jwt issuer identifier
		"my-access-secret",  // jwt.access.secretcode from config
		"my-refresh-secret", // jwt.refresh.secretcode from config
	)
}

// ExampleCreateToken shows how to generate a signed HS256 access token for a user.
// The returned string is a standard Bearer token for use in the Authorization header.
func ExampleCreateToken() {
	tok, err := token.CreateToken(42, "alice", 7, 100, []string{"admin", "editor"}, time.Hour)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(tok != "")
	// Output:
	// true
}

// ExampleParseToken shows the create-and-parse round-trip.
// ParseToken validates the signature, issuer, and token type, then returns
// the embedded LoginUser.
func ExampleParseToken() {
	tok, err := token.CreateToken(42, "alice", 7, 100, []string{"admin", "editor"}, time.Hour)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	user, err := token.ParseToken(tok)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(user.Username)
	fmt.Println(user.UserId)
	fmt.Println(user.OrgId)
	// Output:
	// alice
	// 42
	// 7
}

// ExampleCreateRefreshToken shows how to generate a long-lived refresh token.
// Refresh tokens are signed with a separate secret (jwt.refresh.secretcode)
// and do not carry role information.
func ExampleCreateRefreshToken() {
	tok, err := token.CreateRefreshToken(42, "alice", 7, 100, 24*time.Hour)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(tok != "")
	// Output:
	// true
}

// ExampleParseRefreshToken shows the refresh-token round-trip.
// Note that Roles is always empty in a refresh token.
func ExampleParseRefreshToken() {
	tok, err := token.CreateRefreshToken(42, "alice", 7, 100, 24*time.Hour)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	user, err := token.ParseRefreshToken(tok)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(user.Username)
	fmt.Println(user.OrgId)
	fmt.Println(len(user.Roles) == 0) // refresh tokens carry no roles
	// Output:
	// alice
	// 7
	// true
}

// ExampleAuthenticate shows the Gin middleware in action.
// A valid Bearer token returns 200; a missing header returns 401.
func ExampleAuthenticate() {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(token.Authenticate())
	r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Valid token — 200 OK.
	tok, _ := token.CreateToken(1, "alice", 5, 10, []string{"admin"}, time.Hour)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)
	fmt.Println(w.Code)

	// Missing Authorization header — 401 Unauthorized.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w2, req2)
	fmt.Println(w2.Code)
	// Output:
	// 200
	// 401
}
