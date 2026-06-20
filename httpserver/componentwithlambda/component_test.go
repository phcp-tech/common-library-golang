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
	"net/http"
	"os"
	"testing"

	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/httpserver/componentwithlambda"
)

// TestMain initialises the env singleton with testdata/config.toml.
// The config sets app.runmode = "" (non-lambda) so the plain HTTP path is
// exercised. Port "99999" (out of range) ensures the server goroutine fails
// immediately without occupying a real port.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("httpserver/componentwithlambda tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestComponent_ReturnsNonNil verifies that Component() returns a non-nil IComponent.
func TestComponent_ReturnsNonNil(t *testing.T) {
	c := componentwithlambda.Component(func() http.Handler { return http.NewServeMux() })
	if c == nil {
		t.Error("Component() returned nil")
	}
}

// TestComponent_Name verifies the component name is "httpserver".
func TestComponent_Name(t *testing.T) {
	c := componentwithlambda.Component(func() http.Handler { return http.NewServeMux() })
	if c.Name() != "httpserver" {
		t.Errorf("Component().Name() = %q, want %q", c.Name(), "httpserver")
	}
}

// TestComponent_Init_ReturnsNil verifies that Init() returns nil regardless of
// whether the server goroutine subsequently fails. Init() starts the server
// asynchronously; errors are handled via shutdown.Trigger().
func TestComponent_Init_ReturnsNil(t *testing.T) {
	c := componentwithlambda.Component(func() http.Handler { return http.NewServeMux() })
	if err := c.Init(); err != nil {
		t.Errorf("Component().Init() = %v, want nil (server starts async)", err)
	}
	c.Close()
}
