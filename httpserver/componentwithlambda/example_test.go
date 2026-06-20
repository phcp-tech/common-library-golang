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
