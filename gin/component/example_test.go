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

	"github.com/gin-gonic/gin"
	ginComp "github.com/phcp-tech/common-library-golang/gin/component"
)

// ExampleComponent shows how gin.Component() is registered in a bootstrap chain.
// CORS origins are read from env automatically (cors.allow.origins.prod or .dev).
// The *gin.Engine created inside Init() is shared with the HTTP server component
// via a closure variable in main.
//
//	var router *gin.Engine
//	Add(ginComp.Component(func(r *gin.Engine) {
//	    router = r
//	    adapter.Mount(r)
//	})).
//	Add(httpComp.Component(func() http.Handler { return router }))
func ExampleComponent() {
	var router *gin.Engine
	c := ginComp.Component(func(r *gin.Engine) {
		router = r
	})
	fmt.Println(c != nil)
	fmt.Println(c.Name())
	_ = router
	_ = http.NewServeMux() // suppress unused import
	// Output:
	// true
	// gin
}
