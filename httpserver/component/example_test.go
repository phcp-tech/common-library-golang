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

	httpComp "github.com/phcp-tech/common-library-golang/httpserver/component"
)

// ExampleComponent shows how Component() is used in a bootstrap registration chain.
// It reads app.runmode and http.server.port from env during Init().
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
