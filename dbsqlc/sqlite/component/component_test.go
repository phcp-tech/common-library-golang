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
		panic("sqlite/component tests: failed to load testdata/config.toml: " + err.Error())
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
	if c.Name() != "sqlite" {
		t.Errorf("Component().Name() = %q, want %q", c.Name(), "sqlite")
	}
}

// TestComponent_Init_ReturnsNilWithMemoryDB verifies that Init() succeeds when
// db.sqlite.path is ":memory:".
//
// The in-memory database requires no file system access and always opens
// successfully. sqlite.InitDefault uses sync.Once — this is the sole Init
// call in this test binary.
func TestComponent_Init_ReturnsNilWithMemoryDB(t *testing.T) {
	err := Component().Init()
	if err != nil {
		t.Errorf("Component().Init() = %v, want nil (in-memory database)", err)
	}
}

// TestComponent_Close_WhenDBInitialised verifies that Close() executes the
// full close body (db.Close + slog.Info) without panicking when the default
// SQLite connection is non-nil.
//
// Init() must be called first because sqlite.InitDefault only sets the
// singleton on success. With ":memory:", Init() always succeeds, so
// sqlite.Default() is guaranteed non-nil after this call.
// sqlite.InitDefault uses sync.Once, so the call is idempotent.
func TestComponent_Close_WhenDBInitialised(t *testing.T) {
	c := Component()
	_ = c.Init() // idempotent — ensures sqlite.Default() != nil
	c.Close()    // must not panic; covers db.Close() + slog.Info branch
}
