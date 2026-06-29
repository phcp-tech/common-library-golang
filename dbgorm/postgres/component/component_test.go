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

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dbgorm/postgres"
	"github.com/phcp-tech/common-library-golang/dbgorm/sqlite"
	"github.com/phcp-tech/common-library-golang/env"
)

// TestMain initialises the env singleton once for the entire component test suite.
// loadFromEnv calls env.Env().String(...) which panics if the singleton is nil.
func TestMain(m *testing.M) {
	if err := env.InitEnv("testdata/config.toml"); err != nil {
		panic("dbgorm/postgres/component tests: failed to load testdata/config.toml: " + err.Error())
	}
	os.Exit(m.Run())
}

// TestComponent_ReturnsNonNil verifies that Component() returns a non-nil IComponent.
func TestComponent_ReturnsNonNil(t *testing.T) {
	if Component() == nil {
		t.Error("Component() returned nil")
	}
}

// TestComponent_Name verifies the component name is "postgres".
func TestComponent_Name(t *testing.T) {
	c := Component()
	if c.Name() != "postgres" {
		t.Errorf("Component().Name() = %q, want %q", c.Name(), "postgres")
	}
}

// TestComponent_Init_ReturnsError verifies that Init() propagates the connection
// error when the configured database is unreachable.
//
// GORM's PostgreSQL driver pings the server and runs SHOW search_path on Open —
// port 1 (testdata config) always refuses connections, so Init() returns a
// non-nil error immediately. dbgorm.Default() remains nil because InitDefault
// only calls SetDefault on success.
func TestComponent_Init_ReturnsError(t *testing.T) {
	err := Component().Init()
	if err == nil {
		t.Error("Component().Init() should return an error when the database is unreachable, got nil")
	}
}

// TestComponent_Close_WhenDBNil verifies that Close() does not panic when
// dbgorm.Default() is nil (i.e. Init() failed or was never called).
//
// Because Init() always fails with the testdata config (port 1 unreachable),
// dbgorm.Default() remains nil. The close function exits early via the nil guard.
func TestComponent_Close_WhenDBNil(t *testing.T) {
	c := Component()
	_ = c.Init() // fails — dbgorm.Default() remains nil
	c.Close()    // must not panic; covers the nil-db guard
}

// TestLoadFromEnv_Success covers the slog.Info + return nil path of loadFromEnv.
// postgresInitDefault is stubbed to return nil so the success branch is reached
// without requiring a live PostgreSQL server.
func TestLoadFromEnv_Success(t *testing.T) {
	prev := postgresInitDefault
	postgresInitDefault = func(_ *postgres.Config) error { return nil }
	defer func() { postgresInitDefault = prev }()

	if err := loadFromEnv(); err != nil {
		t.Errorf("loadFromEnv() with stub = %v, want nil", err)
	}
}

// TestComponent_Close_WithLiveDB verifies the non-nil db path of Close()
// by injecting a SQLite in-memory database as the process-wide default.
// This covers the dbgorm.Close(db) + slog.Info branches that are unreachable
// when Init() always fails due to an unreachable PostgreSQL host.
func TestComponent_Close_WithLiveDB(t *testing.T) {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("sqlite.NewSQLite: %v", err)
	}
	prev := dbgorm.Default()
	dbgorm.SetDefault(db)
	defer dbgorm.SetDefault(prev)

	Component().Close() // db != nil → dbgorm.Close succeeds → slog.Info covered
}
