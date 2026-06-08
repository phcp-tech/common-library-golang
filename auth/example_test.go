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

// package auth_test demonstrates the public API from a caller's perspective.
package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/phcp-tech/common-library-golang/auth"
	"github.com/phcp-tech/common-library-golang/dto"

	"github.com/gin-gonic/gin"
)

const exampleModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

const examplePolicy = `
p, admin, /api/data, GET
p, viewer, /api/data, GET
`

// ExampleInitCasbin_strings shows how to initialise the Casbin enforcer from
// in-memory model and policy strings. Pass fs=true to treat the arguments as
// raw strings instead of file paths.
func ExampleInitCasbin_strings() {
	err := auth.InitCasbin(true, exampleModel, examplePolicy)
	fmt.Println(err)
	// Output:
	// <nil>
}

// ExampleInitCasbin_files shows how to initialise the Casbin enforcer from
// model and policy files on disk. Pass fs=false to treat the arguments as
// file paths.
func ExampleInitCasbin_files() {
	err := auth.InitCasbin(false, "model.conf", "policy.csv")
	if err != nil {
		fmt.Println("error:", err != nil) // file not found in test environment
	}
	// Output:
	// error: true
}

// ExampleAuthorize shows how to register the Casbin authorisation middleware
// on a Gin router. The middleware reads roles from the "userInfo" context key
// (set by the token.Authenticate middleware) and enforces the loaded policy.
//
// Method-2 semantics: every role in the user's role list must pass the policy
// check. If any role is denied, the request is rejected with 403 Forbidden.
func ExampleAuthorize() {
	_ = auth.InitCasbin(true, exampleModel, examplePolicy)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Inject userInfo into context (normally done by token.Authenticate).
	r.Use(func(c *gin.Context) {
		c.Set("userInfo", dto.LoginUser{
			UserId:   1,
			Username: "alice",
			Roles:    []string{"admin"},
		})
		c.Next()
	})
	r.Use(auth.Authorize())
	r.GET("/api/data", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Allowed role — 200 OK.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	r.ServeHTTP(w, req)
	fmt.Println(w.Code)
	// Output:
	// 200
}
