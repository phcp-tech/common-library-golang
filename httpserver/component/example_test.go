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

package component_test

import (
	"fmt"
	"net/http"

	"github.com/phcp-tech/common-library-golang/httpserver"
	httpComp "github.com/phcp-tech/common-library-golang/httpserver/component"
)

// ExampleComponent shows how Component() is used in a bootstrap registration chain.
// It reads http.server.port from env during Init(); this package always
// returns a plain HTTP/HTTPS runner (see httpserver/componentwithlambda for
// the sibling that also reads app.runmode to switch to a Lambda runner).
//
// Init() starts the HTTP server in a background goroutine and returns immediately.
// Startup errors (e.g. port already in use) are handled asynchronously via
// shutdown.Trigger(). Close() gracefully drains in-flight requests.
//
// Typical usage with a bridge variable in main:
//
//	var router *gin.Engine
//	Add(ginComp.Component(func(r *gin.Engine) {
//	    router = r
//	    adapter.Mount(r)
//	})).
//	Add(httpComp.Component(func() http.Handler { return router }))
func ExampleComponent() {
	c := httpComp.Component(func() http.Handler { return http.NewServeMux() })
	fmt.Println(c != nil)
	// Output:
	// true
}

// ExampleComponent_withConfig shows the optional Config parameter, for
// services that need to customize the runner (e.g. raising or disabling the
// write timeout for a long-lived SSE/streaming endpoint) without giving up
// bootstrap's lifecycle management.
//
// cfg's Port field is always overwritten from the http.server.port env key
// regardless of what's set here. Pass at most one Config; see [Component]'s
// doc comment for the full contract.
func ExampleComponent_withConfig() {
	c := httpComp.Component(
		func() http.Handler { return http.NewServeMux() },
		httpserver.Config{WriteTimeout: httpserver.NoWriteTimeout},
	)
	fmt.Println(c != nil)
	// Output:
	// true
}

// ExampleComponentWithRunner shows how to supply a custom runner factory.
// Use this when the deployment target determines the runner type at runtime —
// for example, when Lambda support is required (see httpserver/componentwithlambda
// for the ready-made Lambda-aware component).
//
//	httpComp.ComponentWithRunner(
//	    func() http.Handler { return router },
//	    func() httpserver.IRunner {
//	        if isLambda {
//	            return lambda.NewHttpServer()
//	        }
//	        return httpserver.NewHttpServer(httpserver.Config{Port: port})
//	    },
//	)
func ExampleComponentWithRunner() {
	c := httpComp.ComponentWithRunner(
		func() http.Handler { return http.NewServeMux() },
		func() httpserver.IRunner {
			return httpserver.NewHttpServer(httpserver.Config{Port: "8080"})
		},
	)
	fmt.Println(c != nil)
	// Output:
	// true
}
