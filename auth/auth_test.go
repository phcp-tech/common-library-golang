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

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/phcp-tech/common-library-golang/auth"
	"github.com/phcp-tech/common-library-golang/dto"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
)

// Minimal Casbin model and policy used for in-memory testing.
const testCasbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`

const testCasbinPolicy = `
p, admin, /api/test, GET
p, admin, /api/test, POST
p, viewer, /api/test, GET
`

func mustInitCasbin(t *testing.T) {
	t.Helper()
	if err := auth.InitCasbin(true, testCasbinModel, testCasbinPolicy); err != nil {
		t.Fatalf("InitCasbin: %v", err)
	}
}

func newGinWithAuthorize(userInfo dto.LoginUser) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userInfo", userInfo)
		c.Next()
	})
	r.Use(auth.Authorize())
	r.GET("/api/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/api/test", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

// ─── InitCasbin ───────────────────────────────────────────────────────────────

func TestInitCasbin_FromStrings_Succeeds(t *testing.T) {
	err := auth.InitCasbin(true, testCasbinModel, testCasbinPolicy)
	if err != nil {
		t.Errorf("InitCasbin(fs=true) error = %v, want nil", err)
	}
}

func TestInitCasbin_FromStrings_InvalidModel_Panics(t *testing.T) {
	// The source ignores the error from NewModelFromString; a nil model causes
	// casbin.NewEnforcer to panic with "assignment to entry in nil map".
	defer func() {
		if recover() == nil {
			t.Error("expected panic for invalid model string, got none")
		}
	}()
	_ = auth.InitCasbin(true, "not-a-valid-model", "")
}

func TestInitCasbin_FromFiles_NotFound_ReturnsError(t *testing.T) {
	err := auth.InitCasbin(false, "/no/such/model.conf", "/no/such/policy.csv")
	if err == nil {
		t.Error("InitCasbin with non-existent files should return error, got nil")
	}
}

// ─── Authorize middleware ─────────────────────────────────────────────────────

func TestAuthorize_AllowedRole_Returns200(t *testing.T) {
	mustInitCasbin(t)
	r := newGinWithAuthorize(dto.LoginUser{UserId: 1, Username: "alice", Roles: []string{"admin"}})
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthorize_UnknownRole_Returns403(t *testing.T) {
	mustInitCasbin(t)
	r := newGinWithAuthorize(dto.LoginUser{UserId: 2, Username: "bob", Roles: []string{"nobody"}})
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestAuthorize_MultipleRoles_AllMustPass_OneFails_Returns403(t *testing.T) {
	mustInitCasbin(t)
	// viewer has GET but not POST; combined with admin the POST request must still be denied
	// because method-2 logic requires every role to pass.
	r := newGinWithAuthorize(dto.LoginUser{UserId: 3, Username: "charlie", Roles: []string{"admin", "viewer"}})
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("mixed roles POST: status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestAuthorize_EmptyRoles_PassesThrough(t *testing.T) {
	mustInitCasbin(t)
	// Empty roles: the enforcement loop never runs, pass stays true → Next() is called.
	r := newGinWithAuthorize(dto.LoginUser{UserId: 4, Username: "guest", Roles: []string{}})
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("empty roles: status = %d, want %d", w.Code, http.StatusOK)
	}
}

// ─── Mock helpers (merged from casbin_auth_mock.go) ──────────────────────────

// MockAuthorizeWithPass patches auth.Authorize with a pass-through handler that
// calls c.Next() without any permission checks.
func MockAuthorizeWithPass() *gomonkey.Patches {
	return gomonkey.ApplyFunc(auth.Authorize, func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Next()
		}
	})
}

// MockAuthorizeWithNotPass patches auth.Authorize with a handler that always
// responds 403 Forbidden, simulating an authorization failure.
func MockAuthorizeWithNotPass() *gomonkey.Patches {
	return gomonkey.ApplyFunc(auth.Authorize, func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ResponseMessage{Code: http.StatusForbidden, Message: "access forbidden"})
		}
	})
}

// Gomonkey patching requires -gcflags=all=-l to disable inlining; the tests
// below verify the helpers return a valid, resettable Patches object.

func TestMockAuthorizeWithPass_ReturnsValidPatches(t *testing.T) {
	patches := MockAuthorizeWithPass()
	if patches == nil {
		t.Fatal("MockAuthorizeWithPass() returned nil patches")
	}
	patches.Reset() // must not panic
}

func TestMockAuthorizeWithNotPass_ReturnsValidPatches(t *testing.T) {
	patches := MockAuthorizeWithNotPass()
	if patches == nil {
		t.Fatal("MockAuthorizeWithNotPass() returned nil patches")
	}
	patches.Reset() // must not panic
}
