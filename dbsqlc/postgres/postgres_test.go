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

package postgres_test

import (
	"os"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlc/postgres"
	"github.com/phcp-tech/common-library-golang/env"
)

func TestMain(m *testing.M) {
	f, err := os.CreateTemp("", "postgres-test-*.toml")
	if err != nil {
		os.Exit(2)
	}
	f.WriteString("[app]\nname = \"postgres-test\"\n[app.env]\nprefix = \"TEST_\"\n") //nolint:errcheck
	f.Close()
	defer os.Remove(f.Name())

	if err := env.InitEnv(f.Name()); err != nil {
		os.Exit(2)
	}
	os.Exit(m.Run())
}

func TestDefault_BeforeInit_IsNil(t *testing.T) {
	if postgres.Default() != nil {
		t.Skip("singleton already initialised in this process — cannot test pre-init state")
	}
}

// ─── Singleton lifecycle ──────────────────────────────────────────────────────

func TestSingleton_Lifecycle(t *testing.T) {
	if postgres.Default() != nil {
		t.Skip("singleton already initialised before lifecycle test ran")
	}

	// InitDefault with an unreachable host: NewPostgres performs an eager
	// connectivity check (show search_path), so it returns an error immediately.
	err := postgres.InitDefault(&postgres.Config{
		Host:            "127.0.0.1",
		Port:            "19997",
		Database:        "noexist",
		Username:        "nobody",
		Password:        "nopass",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 60,
		ConnMaxIdletime: 10,
	})
	if err == nil {
		t.Fatal("InitDefault should return an error when the database is unreachable")
	}

	// After a failed InitDefault the singleton remains nil.
	if postgres.Default() != nil {
		t.Fatal("Default() should be nil after a failed InitDefault")
	}

	// Second InitDefault is a no-op (sync.Once) — returns nil even though the
	// first call failed; the instance stays nil.
	if err := postgres.InitDefault(&postgres.Config{
		Host:            "127.0.0.1",
		Port:            "19997",
		Database:        "other",
		Username:        "nobody",
		Password:        "nopass",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 60,
		ConnMaxIdletime: 10,
	}); err != nil {
		t.Errorf("second InitDefault should return nil (sync.Once no-op); got %v", err)
	}
}

// ─── NewPostgres — eager connectivity check ───────────────────────────────────

func TestNewPostgres_EagerCheck_ReturnsError(t *testing.T) {
	// NewPostgres issues "show search_path" to verify connectivity.
	// With an unreachable host it returns a non-nil error immediately.
	pool, err := postgres.NewPostgres(&postgres.Config{
		Host:            "127.0.0.1",
		Port:            "19998",
		Database:        "noexist",
		Username:        "nobody",
		Password:        "nopass",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 60,
		ConnMaxIdletime: 10,
	})
	if err == nil {
		pool.Close()
		t.Fatal("NewPostgres should return an error when the database is unreachable")
	}
	if pool != nil {
		t.Fatal("NewPostgres should return nil pool on error")
	}
}
