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
	"time"

	"github.com/phcp-tech/common-library-golang/env"
	"github.com/phcp-tech/common-library-golang/httpserver"
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

// TestComponent_Close_WhenRunnerNil verifies that Close() does not panic when
// called before Init(). The internal runner field is nil in this state, so
// Close() must return immediately via the nil guard without dereferencing it.
func TestComponent_Close_WhenRunnerNil(t *testing.T) {
	c := component.Component(func() http.Handler { return http.NewServeMux() })
	c.Close() // runner == nil — must not panic; covers nil guard + return
}

// TestComponent_Close_AfterInit verifies that Close() executes the full
// shutdown path (context creation, Shutdown call, slog.Info) without panicking.
//
// Init() starts the server goroutine with port "99999" (invalid), which
// causes the goroutine to fail immediately. A brief sleep ensures the
// goroutine has exited before Close() calls Shutdown, avoiding any race
// on the underlying http.Server state.
func TestComponent_Close_AfterInit(t *testing.T) {
	c := component.Component(func() http.Handler { return http.NewServeMux() })
	if err := c.Init(); err != nil {
		t.Fatalf("Init() = %v, want nil", err)
	}
	time.Sleep(20 * time.Millisecond) // let server goroutine fail with invalid port
	c.Close()                         // must not panic; covers ctx/defer/Shutdown/slog.Info
}

// TestComponentWithRunner_UsesProvidedFactory verifies that ComponentWithRunner
// calls the supplied factory function during Init() instead of the default
// loadFromEnv. A custom factory returning a real server on port "0" (OS-assigned)
// is used so the runner is non-nil and Init() succeeds.
func TestComponentWithRunner_UsesProvidedFactory(t *testing.T) {
	var factoryCalled bool
	factory := func() httpserver.IRunner {
		factoryCalled = true
		return httpserver.NewHttpServer(httpserver.Config{Port: "0"})
	}

	c := component.ComponentWithRunner(func() http.Handler { return http.NewServeMux() }, factory)
	if err := c.Init(); err != nil {
		t.Fatalf("ComponentWithRunner().Init() = %v, want nil", err)
	}
	if !factoryCalled {
		t.Error("factory was not called during Init()")
	}
	time.Sleep(20 * time.Millisecond)
	c.Close()
}

// TestComponentWithRunner_ReturnsNonNil verifies that ComponentWithRunner
// returns a non-nil IComponent.
func TestComponentWithRunner_ReturnsNonNil(t *testing.T) {
	c := component.ComponentWithRunner(
		func() http.Handler { return http.NewServeMux() },
		func() httpserver.IRunner { return httpserver.NewHttpServer(httpserver.Config{Port: "0"}) },
	)
	if c == nil {
		t.Error("ComponentWithRunner() returned nil")
	}
}
