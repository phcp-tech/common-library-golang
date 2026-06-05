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

package sqlite_test

import (
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlc/sqlite"
)

// ─── NewSQLite ────────────────────────────────────────────────────────────────

func TestNewSQLite_InMemory(t *testing.T) {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewSQLite(:memory:) error = %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("NewSQLite returned nil db")
	}
	if err := db.Ping(); err != nil {
		t.Errorf("Ping() after NewSQLite = %v, want nil", err)
	}
}

func TestNewSQLite_CanQuery(t *testing.T) {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewSQLite error = %v", err)
	}
	defer db.Close()

	var n int
	if err := db.QueryRow("SELECT 1").Scan(&n); err != nil {
		t.Fatalf("QueryRow(SELECT 1): %v", err)
	}
	if n != 1 {
		t.Errorf("SELECT 1 returned %d, want 1", n)
	}
}

func TestNewSQLite_MultipleConnections(t *testing.T) {
	db1, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("first NewSQLite error = %v", err)
	}
	defer db1.Close()

	db2, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("second NewSQLite error = %v", err)
	}
	defer db2.Close()

	if db1 == db2 {
		t.Error("NewSQLite should return a new handle each call")
	}
}

func TestNewSQLite_Pragmas(t *testing.T) {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewSQLite error = %v", err)
	}
	defer db.Close()

	t.Run("journal_mode", func(t *testing.T) {
		var v string
		if err := db.QueryRow("PRAGMA journal_mode").Scan(&v); err != nil {
			t.Fatalf("PRAGMA journal_mode: %v", err)
		}
		if v != "wal" && v != "memory" {
			t.Errorf("journal_mode = %q, want wal or memory", v)
		}
	})

	t.Run("synchronous", func(t *testing.T) {
		var v int
		if err := db.QueryRow("PRAGMA synchronous").Scan(&v); err != nil {
			t.Fatalf("PRAGMA synchronous: %v", err)
		}
		if v != 1 { // 1 = NORMAL
			t.Errorf("synchronous = %d, want 1 (NORMAL)", v)
		}
	})

	t.Run("foreign_keys", func(t *testing.T) {
		var v int
		if err := db.QueryRow("PRAGMA foreign_keys").Scan(&v); err != nil {
			t.Fatalf("PRAGMA foreign_keys: %v", err)
		}
		if v != 1 {
			t.Errorf("foreign_keys = %d, want 1 (ON)", v)
		}
	})

	t.Run("busy_timeout", func(t *testing.T) {
		var v int
		if err := db.QueryRow("PRAGMA busy_timeout").Scan(&v); err != nil {
			t.Fatalf("PRAGMA busy_timeout: %v", err)
		}
		if v != 5000 {
			t.Errorf("busy_timeout = %d, want 5000", v)
		}
	})

	t.Run("cache_size", func(t *testing.T) {
		var v int
		if err := db.QueryRow("PRAGMA cache_size").Scan(&v); err != nil {
			t.Fatalf("PRAGMA cache_size: %v", err)
		}
		if v != -32000 {
			t.Errorf("cache_size = %d, want -32000", v)
		}
	})
}

func TestNewSQLite_ConnectionPool(t *testing.T) {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewSQLite error = %v", err)
	}
	defer db.Close()

	stats := db.Stats()
	if stats.MaxOpenConnections != 4 {
		t.Errorf("MaxOpenConnections = %d, want 4", stats.MaxOpenConnections)
	}
}

// ─── Singleton lifecycle ──────────────────────────────────────────────────────

func TestSingleton_Lifecycle(t *testing.T) {
	// Before InitDefault: Default() is nil.
	if got := sqlite.Default(); got != nil {
		t.Error("Default() should be nil before InitDefault")
	}

	// InitDefault should succeed with an in-memory database.
	if err := sqlite.InitDefault(&sqlite.Config{Path: ":memory:"}); err != nil {
		t.Fatalf("InitDefault(:memory:) error = %v", err)
	}

	// After init: Default() is non-nil.
	if got := sqlite.Default(); got == nil {
		t.Error("Default() should be non-nil after InitDefault")
	}

	// A second InitDefault call is a no-op (sync.Once).
	firstDB := sqlite.Default()
	if err := sqlite.InitDefault(&sqlite.Config{Path: ":memory:"}); err != nil {
		t.Errorf("second InitDefault should return nil; got %v", err)
	}
	if sqlite.Default() != firstDB {
		t.Error("second InitDefault must not replace the existing instance")
	}
}
