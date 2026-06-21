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

package version_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/version"
)

// ExampleGet shows how to expose build and runtime metadata via a /version
// HTTP endpoint. Requires env.InitEnv to be called at application startup.
func ExampleGet() {
	v := version.Get()
	fmt.Printf("name=%s version=%s env=%s\n", v.Name, v.Version, v.Environment)
}
