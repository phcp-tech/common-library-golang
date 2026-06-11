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

package app_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/app"
)

// ExampleGetHealth shows the typical usage in a /health HTTP endpoint handler.
// Requires env.InitEnv to be called at application startup.
// Status is 2 when the database is reachable, 0 otherwise.
func ExampleGetHealth() {
	h := app.GetHealth()
	if h.Status == 2 {
		fmt.Printf("%s: healthy\n", h.Name)
	} else {
		fmt.Printf("%s: database unreachable\n", h.Name)
	}
}

// ExampleGetVersion shows how to expose build and runtime metadata via a
// /version HTTP endpoint. Requires env.InitEnv to be called at application startup.
func ExampleGetVersion() {
	v := app.GetVersion()
	fmt.Printf("name=%s version=%s env=%s\n", v.Name, v.Version, v.Environment)
}
