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
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/phcp-tech/common-library-golang/dto"
	"github.com/pkg/errors"
)

// slog is used instead of the project log package so that this package has no
// initialisation dependency. slog uses the stdlib default logger when called
// standalone; when the caller invokes log.InitLog(), that function calls
// slog.SetDefault(l.Logger), so all slog calls are automatically routed to the
// project logger with no behaviour change here.

const (
	accessToken  = "access"
	refreshToken = "refresh"
)

var (
	issuer        string
	accessSecret  string
	refreshSecret string
)

// InitToken stores the JWT signing secrets and issuer. It must be called once
// at application startup before any token function.
// The secrets and issuer are typically read from env.Env() after env.InitEnv().
func InitToken(iss, access, refresh string) {
	issuer = iss
	accessSecret = access
	refreshSecret = refresh
}

// UserClaims holds the JWT claims for an authenticated user, including UserId, ProductId,
// and Roles embedded from LoginUser, along with standard registered JWT claims.
// Note: the secret key is assumed trustworthy; field-level encryption may be added in the future.
type UserClaims struct {
	dto.LoginUser
	jwt.RegisteredClaims
}

// CreateToken generates a signed HS256 JWT access token for the given user, valid for the
// specified duration.
func CreateToken(userId int64, username string, orgId int64, productId int64, roles []string, expires time.Duration) (string, error) {
	claims := UserClaims{
		dto.LoginUser{
			OrgId:     orgId,
			ProductId: productId,
			UserId:    userId,
			Username:  username,
			Roles:     roles,
			TokenType: accessToken,
		},
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expires)),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(accessSecret))
}

// ParseToken parses and validates a JWT access token string, returning the embedded
// LoginUser information on success or an error if the token is invalid or expired.
func ParseToken(tokenString string) (userInfo dto.LoginUser, err error) {
	tokenClaims, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(accessSecret), nil
	})
	if err != nil {
		return userInfo, errors.Wrap(err, "parse claims error")
	}

	claims, ok := tokenClaims.Claims.(*UserClaims)
	if !(ok && tokenClaims.Valid) {
		return userInfo, errors.New("couldn't parse claims")
	}

	if claims.Issuer != issuer {
		return userInfo, errors.New("token issuer missed")
	}

	if claims.TokenType != accessToken {
		return userInfo, errors.New("token type is not access")
	}

	userInfo.Token = tokenString
	userInfo.OrgId = claims.OrgId
	userInfo.ProductId = claims.ProductId
	userInfo.UserId = claims.UserId
	userInfo.Username = claims.Username
	userInfo.Roles = claims.Roles
	return userInfo, nil
}

// CreateRefreshToken creates a long-lived refresh token signed with a different
// secret (jwt.refresh.secretcode). The expires parameter controls the token lifetime in minutes.
func CreateRefreshToken(userId int64, username string, orgId int64, productId int64, expires time.Duration) (string, error) {
	claims := UserClaims{
		dto.LoginUser{
			OrgId:     orgId,
			ProductId: productId,
			UserId:    userId,
			Username:  username,
			Roles:     nil, // Refresh token doesn't need roles, it's only used to get a new access token with the same user info.
			TokenType: refreshToken,
		},
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expires)),
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(refreshSecret))
}

// ParseRefreshToken parses and validates a refresh token using the refresh
// secret code. Returns the embedded user info on success, or an error if
// the token is invalid or expired.
func ParseRefreshToken(tokenString string) (dto.LoginUser, error) {
	tokenClaims, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(refreshSecret), nil
	})
	if err != nil {
		return dto.LoginUser{}, errors.Wrap(err, "parse refresh token error")
	}

	claims, ok := tokenClaims.Claims.(*UserClaims)
	if !(ok && tokenClaims.Valid) {
		return dto.LoginUser{}, errors.New("couldn't parse refresh token claims")
	}

	if claims.Issuer != issuer {
		return dto.LoginUser{}, errors.New("refresh token issuer missed")
	}

	if claims.TokenType != refreshToken {
		return dto.LoginUser{}, errors.New("token type is not refresh")
	}

	userInfo := dto.LoginUser{
		OrgId:     claims.OrgId,
		ProductId: claims.ProductId,
		UserId:    claims.UserId,
		Username:  claims.Username,
		Roles:     claims.Roles,
	}
	return userInfo, nil
}

// Authenticate returns a Gin middleware that validates the Bearer JWT token in the
// Authorization header. On success it stores the parsed LoginUser in the request context
// under the key "userInfo"; on failure it aborts with HTTP 401 Unauthorized.
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenstr := c.GetHeader("Authorization")
		kv := strings.Split(tokenstr, " ")
		if len(kv) == 2 && kv[0] == "Bearer" {
			if userInfo, err := ParseToken(kv[1]); err != nil {
				slog.Warn(fmt.Sprintf("JWT authenticate failed, username: %s, uri: %s, method: %s, error: %s", userInfo.Username, c.Request.RequestURI, c.Request.Method, err.Error()))
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ResponseMessage{Code: http.StatusUnauthorized, Message: err.Error()})
				return
			} else {
				//set userInfo to context for next step call
				c.Set("userInfo", userInfo)
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ResponseMessage{Code: http.StatusUnauthorized, Message: "access unauthorized"})
			return
		}

		c.Next()
	}
}
