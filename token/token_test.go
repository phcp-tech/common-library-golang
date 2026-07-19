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

package token

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/phcp-tech/common-library-golang/dto"
)

// ---------------------------------------------------------------------------
// TestMain – bootstrap a minimal koanf environment so env.Env() is non-nil.
// ---------------------------------------------------------------------------

func TestMain(m *testing.M) {
	InitToken("phcp", "test-access-secret-key", "test-refresh-secret-key")
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// CreateToken – basic assertions
// ---------------------------------------------------------------------------

func TestCreateToken_NonEmpty(t *testing.T) {
	tok, err := CreateToken(1, "alice", 5, 100, []string{"admin"}, 60*time.Minute)
	if err != nil {
		t.Fatalf("CreateToken error: %v", err)
	}
	if tok == "" {
		t.Error("CreateToken returned empty string")
	}
}

func TestCreateToken_DifferentUsersProduceDifferentTokens(t *testing.T) {
	tok1, err := CreateToken(1, "alice", 5, 10, []string{"admin"}, 60*time.Minute)
	if err != nil {
		t.Fatalf("CreateToken user1 error: %v", err)
	}
	tok2, err := CreateToken(2, "bob", 6, 20, []string{"user"}, 60*time.Minute)
	if err != nil {
		t.Fatalf("CreateToken user2 error: %v", err)
	}
	if tok1 == tok2 {
		t.Error("different users should produce different JWT tokens")
	}
}

// ---------------------------------------------------------------------------
// CreateToken + ParseToken round-trip
// ---------------------------------------------------------------------------

func TestCreateParseToken_RoundTrip(t *testing.T) {
	cases := []struct {
		name      string
		userId    int64
		username  string
		orgId     int64
		productId int64
		roles     []string
		expires   time.Duration
	}{
		{
			name:      "single role",
			userId:    42,
			username:  "bob",
			orgId:     900,
			productId: 999,
			roles:     []string{"viewer"},
			expires:   60 * time.Minute,
		},
		{
			name:      "multiple roles",
			userId:    7,
			username:  "carol",
			orgId:     201,
			productId: 200,
			roles:     []string{"admin", "editor", "viewer"},
			expires:   30 * time.Minute,
		},
		{
			name:      "zero org and product id",
			userId:    1,
			username:  "dave",
			orgId:     0,
			productId: 0,
			roles:     []string{"user"},
			expires:   120 * time.Minute,
		},
		{
			name:      "empty roles slice",
			userId:    5,
			username:  "eve",
			orgId:     51,
			productId: 50,
			roles:     []string{},
			expires:   10 * time.Minute,
		},
	}

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			tok, err := CreateToken(tc.userId, tc.username, tc.orgId, tc.productId, tc.roles, tc.expires)
			if err != nil {
				t.Fatalf("CreateToken error: %v", err)
			}
			if tok == "" {
				t.Fatal("CreateToken returned empty token")
			}

			info, err := ParseToken(tok)
			if err != nil {
				t.Fatalf("ParseToken error: %v", err)
			}

			if info.UserId != tc.userId {
				t.Errorf("UserId: got %d, want %d", info.UserId, tc.userId)
			}
			if info.Username != tc.username {
				t.Errorf("Username: got %q, want %q", info.Username, tc.username)
			}
			if info.OrgId != tc.orgId {
				t.Errorf("OrgId: got %d, want %d", info.OrgId, tc.orgId)
			}
			if info.ProductId != tc.productId {
				t.Errorf("ProductId: got %d, want %d", info.ProductId, tc.productId)
			}
			if len(info.Roles) != len(tc.roles) {
				t.Errorf("Roles length: got %d, want %d", len(info.Roles), len(tc.roles))
			} else {
				for i, r := range tc.roles {
					if info.Roles[i] != r {
						t.Errorf("Roles[%d]: got %q, want %q", i, info.Roles[i], r)
					}
				}
			}
			// Token field in returned userInfo must be the original token string.
			if info.Token != tok {
				t.Errorf("Token field: got %q, want %q", info.Token, tok)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ParseToken – error paths
// ---------------------------------------------------------------------------

func TestParseToken_InvalidToken(t *testing.T) {
	_, err := ParseToken("this.is.not.a.valid.jwt")
	if err == nil {
		t.Error("ParseToken should return error for invalid token string")
	}
}

func TestParseToken_EmptyToken(t *testing.T) {
	_, err := ParseToken("")
	if err == nil {
		t.Error("ParseToken should return error for empty token string")
	}
}

func TestParseToken_TamperedSignature(t *testing.T) {
	tok, err := CreateToken(1, "user", 1, 1, []string{"role"}, 60*time.Minute)
	if err != nil {
		t.Fatalf("CreateToken error: %v", err)
	}

	// Flip a character in the payload (middle segment) to invalidate the signature.
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatal("token should have 3 segments")
	}
	runes := []rune(parts[1])
	if runes[len(runes)-1] == 'a' {
		runes[len(runes)-1] = 'b'
	} else {
		runes[len(runes)-1] = 'a'
	}
	parts[1] = string(runes)
	_, err = ParseToken(strings.Join(parts, "."))
	if err == nil {
		t.Error("ParseToken should return error for tampered token")
	}
}

func TestParseToken_RefreshTokenRejected(t *testing.T) {
	// A refresh token is signed with a different secret key.
	// ParseToken (which uses the access secret) must reject it.
	refreshTok, err := CreateRefreshToken(1, "user", 1, 1, 60*time.Minute)
	if err != nil {
		t.Fatalf("CreateRefreshToken error: %v", err)
	}
	_, err = ParseToken(refreshTok)
	if err == nil {
		t.Error("ParseToken should fail when given a refresh token (signed with different secret)")
	}
}

// ---------------------------------------------------------------------------
// CreateToken – tokenType "access" is enforced by ParseToken
// ---------------------------------------------------------------------------

func TestCreateToken_TokenTypeIsAccess(t *testing.T) {
	// ParseToken internally validates that tokenType == "access".
	// If it succeeds without error, the type assertion passed.
	tok, err := CreateToken(1, "user", 1, 1, []string{}, 60*time.Minute)
	if err != nil {
		t.Fatalf("CreateToken error: %v", err)
	}
	_, err = ParseToken(tok)
	if err != nil {
		t.Errorf("ParseToken should succeed for access token; got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CreateRefreshToken + ParseRefreshToken round-trip
// ---------------------------------------------------------------------------

func TestCreateParseRefreshToken_RoundTrip(t *testing.T) {
	cases := []struct {
		name      string
		userId    int64
		username  string
		orgId     int64
		productId int64
		expires   time.Duration
	}{
		{
			name:      "standard user",
			userId:    10,
			username:  "frank",
			orgId:     301,
			productId: 300,
			expires:   1440 * time.Minute, // 24 hours in minutes
		},
		{
			name:      "zero user id",
			userId:    0,
			username:  "ghost",
			orgId:     2,
			productId: 1,
			expires:   60 * time.Minute,
		},
	}

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			tok, err := CreateRefreshToken(tc.userId, tc.username, tc.orgId, tc.productId, tc.expires)
			if err != nil {
				t.Fatalf("CreateRefreshToken error: %v", err)
			}
			if tok == "" {
				t.Fatal("CreateRefreshToken returned empty token")
			}

			info, err := ParseRefreshToken(tok)
			if err != nil {
				t.Fatalf("ParseRefreshToken error: %v", err)
			}

			if info.UserId != tc.userId {
				t.Errorf("UserId: got %d, want %d", info.UserId, tc.userId)
			}
			if info.Username != tc.username {
				t.Errorf("Username: got %q, want %q", info.Username, tc.username)
			}
			if info.OrgId != tc.orgId {
				t.Errorf("OrgId: got %d, want %d", info.OrgId, tc.orgId)
			}
			if info.ProductId != tc.productId {
				t.Errorf("ProductId: got %d, want %d", info.ProductId, tc.productId)
			}
			// Refresh token does not carry roles (set to nil in CreateRefreshToken).
			if len(info.Roles) != 0 {
				t.Errorf("Refresh token Roles should be empty, got %v", info.Roles)
			}
		})
	}
}

func TestParseRefreshToken_InvalidToken(t *testing.T) {
	_, err := ParseRefreshToken("bad.token.value")
	if err == nil {
		t.Error("ParseRefreshToken should return error for invalid token")
	}
}

func TestParseRefreshToken_AccessTokenRejected(t *testing.T) {
	// An access token is signed with a different secret.
	// ParseRefreshToken must reject it.
	accessTok, err := CreateToken(1, "user", 1, 1, []string{"role"}, 60*time.Minute)
	if err != nil {
		t.Fatalf("CreateToken error: %v", err)
	}
	_, err = ParseRefreshToken(accessTok)
	if err == nil {
		t.Error("ParseRefreshToken should fail when given an access token")
	}
}

func TestCreateRefreshToken_NonEmpty(t *testing.T) {
	tok, err := CreateRefreshToken(99, "henry", 55, 500, 120*time.Minute)
	if err != nil {
		t.Fatalf("CreateRefreshToken error: %v", err)
	}
	if tok == "" {
		t.Error("CreateRefreshToken returned empty string")
	}
}

// ---------------------------------------------------------------------------
// ParseToken – issuer and tokenType branches
// ---------------------------------------------------------------------------

func makeRawToken(secretKey, issuer, tokenType string) string {
	claims := UserClaims{
		dto.LoginUser{UserId: 1, Username: "test", TokenType: tokenType},
		jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secretKey))
	return tok
}

func TestParseToken_WrongIssuer_ReturnsError(t *testing.T) {
	tok := makeRawToken(accessSecret, "wrong-issuer", "access")
	_, err := ParseToken(tok)
	if err == nil {
		t.Error("ParseToken should return error for wrong issuer")
	}
}

func TestParseToken_WrongType_ReturnsError(t *testing.T) {
	tok := makeRawToken(accessSecret, issuer, "refresh")
	_, err := ParseToken(tok)
	if err == nil {
		t.Error("ParseToken should return error when tokenType is not access")
	}
}

func TestParseRefreshToken_WrongIssuer_ReturnsError(t *testing.T) {
	tok := makeRawToken(refreshSecret, "bad-issuer", "refresh")
	_, err := ParseRefreshToken(tok)
	if err == nil {
		t.Error("ParseRefreshToken should return error for wrong issuer")
	}
}

func TestParseRefreshToken_WrongType_ReturnsError(t *testing.T) {
	tok := makeRawToken(refreshSecret, issuer, "access")
	_, err := ParseRefreshToken(tok)
	if err == nil {
		t.Error("ParseRefreshToken should return error when tokenType is not refresh")
	}
}

// ---------------------------------------------------------------------------
// Authenticate middleware
// ---------------------------------------------------------------------------

func newAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Authenticate())
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestAuthenticate_ValidToken_ReturnsOK(t *testing.T) {
	tok, err := CreateToken(1, "alice", 5, 10, []string{"admin"}, 60*time.Minute)
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}

	r := newAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthenticate_NoHeader_Returns401(t *testing.T) {
	r := newAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticate_MalformedHeader_Returns401(t *testing.T) {
	r := newAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthenticate_InvalidToken_Returns401(t *testing.T) {
	r := newAuthRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.valid")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// IsUsableSecret — the single source of truth for "is this a real secret",
// shared by CreateToken/ParseToken/CreateRefreshToken/ParseRefreshToken below
// and by token/component's own config validation (token.IsUsableSecret).
// ---------------------------------------------------------------------------

func TestIsUsableSecret(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"placeholder", PlaceholderSecret, false},
		{"real secret", "a-real-random-secret", true},
	}
	for _, c := range cases {
		if got := IsUsableSecret(c.in); got != c.want {
			t.Errorf("IsUsableSecret(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Uninitialised secrets — CreateToken/ParseToken/CreateRefreshToken/
// ParseRefreshToken must refuse to run against an empty accessSecret/
// refreshSecret rather than silently signing/validating with an empty HMAC
// key (which succeeds cryptographically and would otherwise mask a missing
// InitToken call, e.g. a caller that never wired up token/component).
// ---------------------------------------------------------------------------

func TestCreateToken_EmptyAccessSecret_ReturnsError(t *testing.T) {
	savedAccess := accessSecret
	accessSecret = ""
	defer func() { accessSecret = savedAccess }()

	if _, err := CreateToken(1, "alice", 5, 100, []string{"admin"}, time.Hour); err != errAccessSecretNotInitialized {
		t.Errorf("CreateToken() error = %v, want %v", err, errAccessSecretNotInitialized)
	}
}

func TestParseToken_EmptyAccessSecret_ReturnsError(t *testing.T) {
	tok, err := CreateToken(1, "alice", 5, 100, []string{"admin"}, time.Hour)
	if err != nil {
		t.Fatalf("CreateToken() unexpected error: %v", err)
	}

	savedAccess := accessSecret
	accessSecret = ""
	defer func() { accessSecret = savedAccess }()

	if _, err := ParseToken(tok); err != errAccessSecretNotInitialized {
		t.Errorf("ParseToken() error = %v, want %v", err, errAccessSecretNotInitialized)
	}
}

func TestCreateRefreshToken_EmptyRefreshSecret_ReturnsError(t *testing.T) {
	savedRefresh := refreshSecret
	refreshSecret = ""
	defer func() { refreshSecret = savedRefresh }()

	if _, err := CreateRefreshToken(1, "alice", 5, 100, time.Hour); err != errRefreshSecretNotInitialized {
		t.Errorf("CreateRefreshToken() error = %v, want %v", err, errRefreshSecretNotInitialized)
	}
}

func TestParseRefreshToken_EmptyRefreshSecret_ReturnsError(t *testing.T) {
	tok, err := CreateRefreshToken(1, "alice", 5, 100, time.Hour)
	if err != nil {
		t.Fatalf("CreateRefreshToken() unexpected error: %v", err)
	}

	savedRefresh := refreshSecret
	refreshSecret = ""
	defer func() { refreshSecret = savedRefresh }()

	if _, err := ParseRefreshToken(tok); err != errRefreshSecretNotInitialized {
		t.Errorf("ParseRefreshToken() error = %v, want %v", err, errRefreshSecretNotInitialized)
	}
}

// TestInitToken_PlaceholderSecrets_StillRejected reproduces calling InitToken
// directly (bypassing token/component) with the workspace's scaffolding
// placeholder value left un-replaced — the exact scenario a service reaches
// if it never wires up token/component and never overrides jwt.access/
// refresh.secretcode via env. All four functions must still refuse to run,
// not just when the secret is empty.
func TestInitToken_PlaceholderSecrets_StillRejected(t *testing.T) {
	savedIssuer, savedAccess, savedRefresh := issuer, accessSecret, refreshSecret
	defer func() { issuer, accessSecret, refreshSecret = savedIssuer, savedAccess, savedRefresh }()

	InitToken("phcp", PlaceholderSecret, PlaceholderSecret)

	if _, err := CreateToken(1, "alice", 5, 100, []string{"admin"}, time.Hour); err != errAccessSecretNotInitialized {
		t.Errorf("CreateToken() error = %v, want %v", err, errAccessSecretNotInitialized)
	}
	if _, err := ParseToken("irrelevant"); err != errAccessSecretNotInitialized {
		t.Errorf("ParseToken() error = %v, want %v", err, errAccessSecretNotInitialized)
	}
	if _, err := CreateRefreshToken(1, "alice", 5, 100, time.Hour); err != errRefreshSecretNotInitialized {
		t.Errorf("CreateRefreshToken() error = %v, want %v", err, errRefreshSecretNotInitialized)
	}
	if _, err := ParseRefreshToken("irrelevant"); err != errRefreshSecretNotInitialized {
		t.Errorf("ParseRefreshToken() error = %v, want %v", err, errRefreshSecretNotInitialized)
	}
}
