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

// Package component provides Gin lifecycle integration for bootstrap.
package component

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/env"
	libGin "github.com/phcp-tech/common-library-golang/gin"
)

// Compile-time check: ginComponent implements bootstrap.IComponent.
var _ bootstrap.IComponent = (*ginComponent)(nil)

type ginComponent struct {
	mount func(*gin.Engine)
}

func (g *ginComponent) Name() string { return "gin" }

func (g *ginComponent) Init() error {
	// Read CORS origins from env based on the current environment.
	// Prod uses cors.allow.origins.prod; all other environments use cors.allow.origins.dev.
	var origins []string
	if strings.EqualFold(env.Env().String("app.env.value"), "prod") {
		origins = env.Env().Strings("cors.allow.origins.prod")
	} else {
		origins = env.Env().Strings("cors.allow.origins.dev")
	}

	router := libGin.InitGin(origins)
	if g.mount != nil {
		g.mount(router)
	}
	return nil
}

// Close is a no-op: the Gin router has no resources to release; its lifecycle
// is managed by the HTTP server component.
func (g *ginComponent) Close() {}

// Component initialises the Gin router and mounts all routes.
//
// CORS origins are read from env automatically during Init():
//   - prod environment: cors.allow.origins.prod
//   - other environments: cors.allow.origins.dev
//
// mount is called once during Init() to register all routes. The *gin.Engine
// created here should be shared with the HTTP server component via a closure
// variable captured in the caller's main function:
//
//	var router *gin.Engine
//	Add(ginComp.Component(func(r *gin.Engine) {
//	    router = r
//	    adapter.Mount(r)
//	})).
//	Add(httpComp.Component(func() http.Handler { return router }))
func Component(mount func(*gin.Engine)) bootstrap.IComponent {
	return &ginComponent{mount: mount}
}
