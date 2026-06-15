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

package component

import (
	"os"
	"testing"

	"github.com/phcp-tech/common-library-golang/env"
)

// TestMain initialises the env singleton once for the entire component test suite.
// loadFromEnv calls env.Env().String(...) which panics if the singleton is nil.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("postgres/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestComponent_ReturnsNonNil verifies that Component() returns a non-nil IComponent.
func TestComponent_ReturnsNonNil(t *testing.T) {
	c := Component()
	if c == nil {
		t.Error("Component() returned nil")
	}
}

// TestComponent_Name verifies the component name.
func TestComponent_Name(t *testing.T) {
	c := Component()
	if c.Name() != "postgres" {
		t.Errorf("Component().Name() = %q, want %q", c.Name(), "postgres")
	}
}

// TestComponent_Init_ReturnsErrorWhenDatabaseUnreachable verifies that Init
// propagates the connection error when the configured database is unreachable.
//
// The testdata config points to 127.0.0.1:1 which always refuses connections.
// NewPostgres performs an eager connectivity check (show search_path) so the
// connection failure is returned immediately rather than discovered on first query.
//
// postgres.InitDefault uses sync.Once — this is the sole Init call in this
// test binary. The success path requires a live PostgreSQL integration test.
func TestComponent_Init_ReturnsErrorWhenDatabaseUnreachable(t *testing.T) {
	err := Component().Init()
	if err == nil {
		t.Error("Component().Init() should return an error when the database is unreachable, got nil")
	}
}
