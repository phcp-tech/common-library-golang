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

package mysql_test

import (
	"os"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlc/mysql"
	"github.com/phcp-tech/common-library-golang/env"
)

func TestMain(m *testing.M) {
	f, err := os.CreateTemp("", "mysql-test-*.toml")
	if err != nil {
		os.Exit(2)
	}
	f.WriteString("[app]\nname = \"mysql-test\"\n[app.env]\nprefix = \"TEST_\"\n") //nolint:errcheck
	f.Close()
	defer os.Remove(f.Name())

	if err := env.InitEnv(f.Name()); err != nil {
		os.Exit(2)
	}
	os.Exit(m.Run())
}

func TestDefault_BeforeInit_IsNil(t *testing.T) {
	if mysql.Default() != nil {
		t.Skip("singleton already initialised in this process — cannot test pre-init state")
	}
}

// ─── Singleton lifecycle ──────────────────────────────────────────────────────

func TestSingleton_Lifecycle(t *testing.T) {
	if mysql.Default() != nil {
		t.Skip("singleton already initialised before lifecycle test ran")
	}

	// InitDefault with an unreachable host: sql.Open succeeds (lazy connection).
	err := mysql.InitDefault(&mysql.Config{
		Host:     "127.0.0.1",
		Port:     "13307",
		Database: "testdb",
		Username: "nobody",
		Password: "nopass",
	})
	if err != nil {
		t.Fatalf("InitDefault error = %v (sql.Open should not connect)", err)
	}

	if mysql.Default() == nil {
		t.Fatal("Default() is nil after successful InitDefault")
	}

	// Second InitDefault is a no-op (sync.Once).
	if err := mysql.InitDefault(&mysql.Config{
		Host:     "127.0.0.1",
		Port:     "13307",
		Database: "other",
		Username: "nobody",
		Password: "nopass",
	}); err != nil {
		t.Errorf("second InitDefault should return nil; got %v", err)
	}
}

// ─── NewMySQL — sql.Open is lazy, no live server required ────────────────────

func TestNewMySQL_OpenIsLazy(t *testing.T) {
	db, err := mysql.NewMySQL(&mysql.Config{
		Host:            "127.0.0.1",
		Port:            "13307",
		Database:        "testdb",
		Username:        "nobody",
		Password:        "nopass",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 60,
		ConnMaxIdletime: 10,
	})
	if err != nil {
		t.Fatalf("NewMySQL error = %v (sql.Open should not connect)", err)
	}
	if db == nil {
		t.Fatal("NewMySQL returned nil *sql.DB")
	}
	db.Close()
}
