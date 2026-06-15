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
	"net/http"
	"os"
	"testing"

	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/httpserver/component"
)

// TestMain initialises the env singleton with testdata/config.toml.
// The config sets port "99999" (out of range) so the server goroutine fails
// immediately, exercising the error path without occupying a real port.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("httpserver/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestComponent_ReturnsNonNil verifies that Component() returns a non-nil IComponent.
func TestComponent_ReturnsNonNil(t *testing.T) {
	c := component.Component(func() http.Handler { return http.NewServeMux() })
	if c == nil {
		t.Error("Component() returned nil")
	}
}

// TestComponent_Name verifies the component name.
func TestComponent_Name(t *testing.T) {
	c := component.Component(func() http.Handler { return http.NewServeMux() })
	if c.Name() != "httpserver" {
		t.Errorf("Component().Name() = %q, want %q", c.Name(), "httpserver")
	}
}

// TestComponent_Init_ReturnsNil verifies that Init() always returns nil.
//
// Init() starts the server in a background goroutine and returns immediately.
// With port "99999" (out of range) the goroutine's Start() call fails, but
// this error is handled asynchronously via shutdown.Trigger(); Init() itself
// does not block waiting for the server to be ready.
func TestComponent_Init_ReturnsNil(t *testing.T) {
	c := component.Component(func() http.Handler { return http.NewServeMux() })
	err := c.Init()
	if err != nil {
		t.Errorf("Component().Init() = %v, want nil (server starts async)", err)
	}
}

// TestComponent_ReturnType verifies the compile-time return type.
func TestComponent_ReturnType(t *testing.T) {
	c := component.Component(func() http.Handler { return http.NewServeMux() })
	var _ interface {
		Name() string
		Init() error
		Close()
	} = c
}
