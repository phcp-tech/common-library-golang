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

package auth

import (
	"log/slog"
	"net/http"

	"github.com/phcp-tech/common-library-golang/dto"

	"github.com/casbin/casbin/v2"
	casbinModel "github.com/casbin/casbin/v2/model"
	casbinAdapter "github.com/casbin/casbin/v2/persist/string-adapter"
	"github.com/gin-gonic/gin"
)

// slog is used instead of the project log package so that this package has no
// initialisation dependency. slog uses the stdlib default logger when called
// standalone; when the caller invokes log.InitLog(), that function calls
// slog.SetDefault(l.Logger), so all slog calls are automatically routed to the
// project logger with no behaviour change here.

var instance *casbin.Enforcer

// InitCasbin initializes the global Casbin enforcer from either in-memory strings or file paths.
// When fs is true, configModel and configPolicy are treated as raw model/policy strings;
// otherwise they are treated as file paths to the model and policy files.
func InitCasbin(fs bool, configModel string, configPolicy string) error {
	if fs {
		modelString, _ := casbinModel.NewModelFromString(configModel)
		policyString := casbinAdapter.NewAdapter(configPolicy)

		var err error
		instance, err = casbin.NewEnforcer(modelString, policyString)
		return err
	}

	var err error
	instance, err = casbin.NewEnforcer(configModel, configPolicy)
	return err
}

// Authorize returns a Gin middleware that enforces Casbin policy for the current request.
// It reads the authenticated user's roles from the "userInfo" context key and verifies
// that every role is permitted to access the requested URI and HTTP method.
// If any role fails the check the request is rejected with HTTP 403 Forbidden.
func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pass bool = true

		//Method 1:Roles组中有一个Role通过就可以获得授权，废弃使用。//var pass bool = false
		//Method 2:Roles组中有一个Role不通过就不可以获得授权。 //var pass bool = true
		userInfo := c.MustGet("userInfo").(dto.LoginUser)
		for _, role := range userInfo.Roles {
			if result, _ := instance.Enforce(role, c.Request.RequestURI, c.Request.Method); !result {
				pass = false
				break
			}
		}

		if pass {
			c.Next()
		} else {
			// Casbin authorization failed, log the details for debugging
			slog.Warn("Casbin authorization failed",
				"username", userInfo.Username,
				"uri", c.Request.RequestURI,
				"method", c.Request.Method,
			)
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ResponseMessage{Code: http.StatusForbidden, Message: "access forbidden"})
		}
	}
}
