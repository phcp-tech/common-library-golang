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

	authComp "github.com/phcp-tech/common-library-golang/auth/component"
)

// ExampleComponent shows how auth.Component() is registered in a bootstrap chain.
// When fromString is true, model and policy are loaded from the embedded strings
// (suitable for single-binary deployments).
func ExampleComponent() {
	c := authComp.Component(true, testModel, testPolicy)
	fmt.Println(c != nil)
	fmt.Println(c.Name())
	// Output:
	// true
	// casbin
}
