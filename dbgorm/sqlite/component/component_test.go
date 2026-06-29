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
	"errors"
	"os"
	"testing"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dbgorm/sqlite"
	"github.com/phcp-tech/common-library-golang/env"
)

// TestMain initialises the env singleton once for the entire component test suite.
// loadFromEnv calls env.Env().String(...) which panics if the singleton is nil.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("dbgorm/sqlite/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestComponent_ReturnsNonNil verifies that Component() returns a non-nil IComponent.
func TestComponent_ReturnsNonNil(t *testing.T) {
	if Component() == nil {
		t.Error("Component() returned nil")
	}
}

// TestComponent_Name verifies the component name is "sqlite".
func TestComponent_Name(t *testing.T) {
	c := Component()
	if c.Name() != "sqlite" {
		t.Errorf("Component().Name() = %q, want %q", c.Name(), "sqlite")
	}
}

// TestComponent_Init_ReturnsNil verifies that Init() returns nil.
//
// SQLite is an embedded database — no network connection is attempted.
// The testdata config uses an in-memory path ("file::memory:?cache=shared")
// which always succeeds. Init() fails only when the path is empty.
func TestComponent_Init_ReturnsNil(t *testing.T) {
	err := Component().Init()
	if err != nil {
		t.Errorf("Component().Init() = %v, want nil", err)
	}
}

// TestComponent_Close_AfterInit verifies that Close() executes without panicking
// after a successful Init().
//
// Because Init() succeeds with the in-memory path, dbgorm.Default() is non-nil,
// so Close() exercises the full dbgorm.Close() and slog.Info path.
func TestComponent_Close_AfterInit(t *testing.T) {
	c := Component()
	_ = c.Init() // no-op if already set; dbgorm.Default() is non-nil
	c.Close()    // must not panic; exercises dbgorm.Close + slog.Info path
}

// TestLoadFromEnv_Error covers the slog.Error + return err path of loadFromEnv.
// sqliteInitDefault is stubbed to return an error so the error branch is reached
// without requiring db.sqlite.path to be empty in the env configuration.
func TestLoadFromEnv_Error(t *testing.T) {
	prev := sqliteInitDefault
	sqliteInitDefault = func(_ *sqlite.Config) error {
		return errors.New("stub: InitDefault failed")
	}
	defer func() { sqliteInitDefault = prev }()

	if err := loadFromEnv(); err == nil {
		t.Error("loadFromEnv() with failing stub: want error, got nil")
	}
}

// TestComponent_Close_WhenDBNil covers the nil guard return in the close closure.
// dbgorm.Default() is set to nil so the close function exits early.
func TestComponent_Close_WhenDBNil(t *testing.T) {
	prev := dbgorm.Default()
	dbgorm.SetDefault(nil)
	defer dbgorm.SetDefault(prev)

	Component().Close() // db == nil → return immediately, no panic
}
