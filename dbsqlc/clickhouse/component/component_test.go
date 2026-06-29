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

	"github.com/phcp-tech/common-library-golang/dbsqlc/clickhouse"
	"github.com/phcp-tech/common-library-golang/env"
)

func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("dbsqlc/clickhouse/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestComponent_ReturnsNonNil verifies that Component() returns a non-nil IComponent.
func TestComponent_ReturnsNonNil(t *testing.T) {
	if Component() == nil {
		t.Error("Component() returned nil")
	}
}

// TestComponent_Name verifies the component name is "clickhouse".
func TestComponent_Name(t *testing.T) {
	if c := Component(); c.Name() != "clickhouse" {
		t.Errorf("Component().Name() = %q, want %q", c.Name(), "clickhouse")
	}
}

// TestComponent_Close_WhenConnNil verifies that Close() does not panic when
// Default() is nil (before any successful Init call).
func TestComponent_Close_WhenConnNil(t *testing.T) {
	// Default() is nil at this point — no real InitDefault has been called yet.
	c := Component()
	c.Close() // must not panic; covers the nil-conn guard
}

// TestLoadFromEnv_Error covers the slog.Error + return err path of loadFromEnv.
// clickhouseInitDefault is stubbed to return an error.
func TestLoadFromEnv_Error(t *testing.T) {
	prev := clickhouseInitDefault
	clickhouseInitDefault = func(_ *clickhouse.Config) error {
		return errors.New("stub: connect failed")
	}
	defer func() { clickhouseInitDefault = prev }()

	if err := loadFromEnv(); err == nil {
		t.Error("loadFromEnv() with failing stub: want error, got nil")
	}
}

// TestLoadFromEnv_Success covers the slog.Info + return nil path of loadFromEnv.
// clickhouseInitDefault is stubbed to return nil.
func TestLoadFromEnv_Success(t *testing.T) {
	prev := clickhouseInitDefault
	clickhouseInitDefault = func(_ *clickhouse.Config) error { return nil }
	defer func() { clickhouseInitDefault = prev }()

	if err := loadFromEnv(); err != nil {
		t.Errorf("loadFromEnv() with nil stub = %v, want nil", err)
	}
}

// TestComponent_Init_ReturnsNil verifies that Init() always returns nil.
//
// clickhouse-go opens connections lazily — the TCP connection is not established
// at Open time, so InitDefault always succeeds even with an unreachable host.
// After this test, clickhouse.Default() is non-nil (singleton set by InitDefault).
func TestComponent_Init_ReturnsNil(t *testing.T) {
	err := Component().Init()
	if err != nil {
		t.Errorf("Component().Init() = %v, want nil (lazy connection)", err)
	}
}

// TestComponent_Close_AfterInit verifies that Close() executes the full shutdown
// path without panicking after a successful Init().
//
// Because Init() succeeds with a lazy driver, clickhouse.Default() is non-nil,
// so Close() exercises conn.Close() and the slog.Info path.
func TestComponent_Close_AfterInit(t *testing.T) {
	c := Component()
	_ = c.Init() // no-op — singleton already set by TestComponent_Init_ReturnsNil
	c.Close()    // must not panic; covers conn.Close() + slog.Info path
}
