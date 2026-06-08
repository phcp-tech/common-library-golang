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

	// InitDefault with an unreachable host: pgxpool.NewWithConfig is lazy, succeeds.
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
	if err != nil {
		t.Fatalf("InitDefault error = %v (pool creation should not require a live server)", err)
	}

	if postgres.Default() == nil {
		t.Fatal("Default() is nil after successful InitDefault")
	}

	// Second InitDefault is a no-op (sync.Once).
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
		t.Errorf("second InitDefault should return nil; got %v", err)
	}
}

// ─── NewPostgres — pool creation is lazy, no live server required ────────────

func TestNewPostgres_LazyPool(t *testing.T) {
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
	if err != nil {
		t.Fatalf("NewPostgres error = %v (pool creation should not require a live server)", err)
	}
	if pool == nil {
		t.Fatal("NewPostgres returned nil pool")
	}
	pool.Close()
}
