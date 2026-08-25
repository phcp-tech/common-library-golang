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

package componentwithlambda_test

import (
	"fmt"
	"net/http"

	"github.com/phcp-tech/common-library-golang/httpserver"
	httpComp "github.com/phcp-tech/common-library-golang/httpserver/componentwithlambda"
)

// ExampleComponent shows how Component() is used in a bootstrap registration chain.
//
// The runner is selected at Init() time based on the app.runmode env key:
//   - "aws_lambda" → AWS Lambda runner (httpserver/lambda)
//   - anything else → plain HTTP server (httpserver)
//
// Import this package instead of httpserver/component when the service may be
// deployed on AWS Lambda. It pulls in the AWS Lambda SDK, so services that
// never run on Lambda should use httpserver/component to avoid the extra
// dependency.
//
//	var router *gin.Engine
//	bootstrap.New().
//	    Add(envComp.Component("config/app.toml", &configFS)). // 1st — env
//	    Add(logComp.Component()).                              // 2nd — log
//	    Add(ginComp.Component(func(r *gin.Engine) {
//	        router = r
//	        adapter.Mount(r)
//	    })).
//	    Add(httpComp.Component(func() http.Handler { return router })).
//	    Run()
func ExampleComponent() {
	c := httpComp.Component(func() http.Handler { return http.NewServeMux() })
	fmt.Println(c != nil)
	// Output:
	// true
}

// ExampleComponent_withConfig shows the optional Config parameter, for
// services that need to customize the VM-mode runner (e.g. raising or
// disabling the write timeout for a long-lived SSE/streaming endpoint)
// without losing Lambda support.
//
// cfg only applies to the VM-mode branch: its Port field is always
// overwritten from the http.server.port env key regardless of what's set
// here, and Lambda mode ignores cfg entirely (AWS controls that runtime's
// lifecycle, not this package). Pass at most one Config; see [Component]'s
// doc comment for the full contract.
//
//	Add(httpComp.Component(func() http.Handler { return router },
//	    httpserver.Config{WriteTimeout: httpserver.NoWriteTimeout})).
func ExampleComponent_withConfig() {
	c := httpComp.Component(
		func() http.Handler { return http.NewServeMux() },
		httpserver.Config{WriteTimeout: httpserver.NoWriteTimeout},
	)
	fmt.Println(c != nil)
	// Output:
	// true
}
