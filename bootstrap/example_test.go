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

package bootstrap_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/bootstrap"
)

// ExampleFunc shows how to create a simple IComponent from a pair of functions.
// close may be nil when no cleanup is needed.
func ExampleFunc() {
	c := bootstrap.Func("my-service",
		func() error {
			// initialise service
			return nil
		},
		func() {
			// release resources
		},
	)
	fmt.Println(c.Name())
	fmt.Println(c.Init())
	// Output:
	// my-service
	// <nil>
}
