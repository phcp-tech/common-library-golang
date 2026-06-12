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

package loader_test

import (
	"os"
	"testing"

	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/httpserver"
	"github.com/phcp-tech/common-library-golang/httpserver/loader"
)

// TestMain initialises the env singleton with testdata/config.toml.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("httpserver/loader tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestLoadFromEnv_ReturnsNonNilRunner verifies that LoadFromEnv always returns
// a non-nil Runner. NewHttpServer does not bind the port — binding happens
// inside Start — so an invalid port in config is not an error at this stage.
func TestLoadFromEnv_ReturnsNonNilRunner(t *testing.T) {
	runner := loader.LoadFromEnv()
	if runner == nil {
		t.Error("LoadFromEnv() returned nil runner")
	}
}

// TestLoadFromEnv_ReturnType verifies the compile-time return type.
func TestLoadFromEnv_ReturnType(t *testing.T) {
	var fn func() httpserver.Runner = loader.LoadFromEnv
	_ = fn
}
