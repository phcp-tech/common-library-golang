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
		panic("mysql/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestComponent_ReturnsNonNil verifies that Component() returns a non-nil IComponent.
func TestComponent_ReturnsNonNil(t *testing.T) {
	if Component() == nil {
		t.Error("Component() returned nil")
	}
}

// TestComponent_Name verifies the component name is "mysql".
func TestComponent_Name(t *testing.T) {
	c := Component()
	if c.Name() != "mysql" {
		t.Errorf("Component().Name() = %q, want %q", c.Name(), "mysql")
	}
}

// TestComponent_Init_ReturnsNil verifies that Init() always returns nil.
//
// Unlike PostgreSQL (which performs an eager connectivity check), MySQL's
// sql.Open is lazy — no connection is established during Init(), so Init()
// succeeds even when the configured database is unreachable (port 13307).
// mysql.InitDefault uses sync.Once — this is the sole Init call in this binary.
func TestComponent_Init_ReturnsNil(t *testing.T) {
	err := Component().Init()
	if err != nil {
		t.Errorf("Component().Init() = %v, want nil (MySQL sql.Open is lazy)", err)
	}
}

// TestComponent_Close_AfterInit verifies that Close() executes db.Close()
// without panicking after a successful Init().
//
// Because Init() always succeeds with a lazy driver, mysql.Default() is
// non-nil at this point, so Close() exercises the full db.Close() path.
func TestComponent_Close_AfterInit(t *testing.T) {
	c := Component()
	_ = c.Init() // no-op: singleton already initialised by TestComponent_Init_ReturnsNil
	c.Close()    // must not panic; covers db.Close() + slog.Info path
}
