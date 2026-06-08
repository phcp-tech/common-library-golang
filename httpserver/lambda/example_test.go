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

// package lambda_test demonstrates the public API from a caller's perspective.
// For a guide on when to use this package versus the parent httpserver package,
// see the httpserver package examples.
package lambda_test

import (
	"fmt"

	lambdarunner "github.com/phcp-tech/common-library-golang/httpserver/lambda"
)

// ExampleNewHttpServer shows how to create a Lambda runner.
// Import this sub-package only for Lambda deployments; it pulls in the AWS SDK.
// For plain HTTP/HTTPS servers use the parent httpserver package instead.
func ExampleNewHttpServer() {
	runner := lambdarunner.NewHttpServer()
	fmt.Println(runner != nil)
	// Output:
	// true
}
