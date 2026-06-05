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

// package env_test demonstrates the public API from a caller's perspective.
package env_test

import (
	"embed"
	"fmt"
	"log"

	"github.com/phcp-tech/common-library-golang/env"
)

// testdata/config.toml is embedded to demonstrate the embedded-FS code path.
//
//go:embed testdata/config.toml
var exampleFS embed.FS

// Example demonstrates the complete lifecycle: call [InitEnv] once at
// application startup with a TOML config file, then use [Env] to read
// typed values anywhere in the application.
func Example() {
	// Initialise once – typically in main() before starting any services.
	if err := env.InitEnv("/etc/myapp/config.toml"); err != nil {
		log.Fatal(err)
	}

	// Env() returns the loaded configuration; use koanf's typed accessors.
	host := env.Env().String("server.host")
	port := env.Env().Int("server.port")
	fmt.Printf("server: %s:%d\n", host, port)
}

// ExampleInitEnv_embeddedFS shows how to bundle the config file inside the
// binary at compile time using //go:embed, so no external file is needed at
// runtime. Pass the embedded.FS as the second argument to InitEnv.
func ExampleInitEnv_embeddedFS() {
	if err := env.InitEnv("testdata/config.toml", &exampleFS); err != nil {
		log.Fatal(err)
	}

	fmt.Println(env.Env().String("server.host.name"))
	fmt.Println(env.Env().String("server.host.ip"))
	fmt.Println(env.Env().Int("server.port"))
}

// ExampleInitEnv_missingFile shows that InitEnv returns a non-nil error when
// the config file does not exist. The application should treat this as fatal.
func ExampleInitEnv_missingFile() {
	err := env.InitEnv("/nonexistent/config.toml")
	fmt.Println(err != nil)
	// Output: true
}
