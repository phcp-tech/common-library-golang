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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dbsqlx/sqlite"
)

func TestNewSQLite_ErrMissingConfig(t *testing.T) {
	_, err := sqlite.NewSQLite(nil)
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("NewSQLite(nil): err = %v, want ErrMissingConfig", err)
	}
}

func TestNewSQLite_ErrMissingPath(t *testing.T) {
	_, err := sqlite.NewSQLite(&sqlite.Config{})
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("NewSQLite empty path: err = %v, want ErrMissingConfig", err)
	}
}

func TestNewSQLite_InMemory(t *testing.T) {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewSQLite in-memory: %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	if err := db.Ping(); err != nil {
		t.Errorf("Ping in-memory db: %v", err)
	}
}

func TestNewSQLite_AppliesPragmas(t *testing.T) {
	// WAL mode requires a file-backed database — SQLite silently keeps
	// ":memory:" databases on journal_mode=memory regardless of the PRAGMA.
	path := filepath.Join(t.TempDir(), "pragmas.db")
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: path})
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	var journalMode string
	if err := db.Get(&journalMode, "PRAGMA journal_mode"); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
	}

	var foreignKeys int
	if err := db.Get(&foreignKeys, "PRAGMA foreign_keys"); err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("foreign_keys = %d, want 1", foreignKeys)
	}
}

func TestNewSQLite_CreatesParentDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "level1", "level2")
	path := filepath.Join(dir, "app.db")

	db, err := sqlite.NewSQLite(&sqlite.Config{Path: path})
	if err != nil {
		t.Fatalf("NewSQLite with nested path: %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	if _, statErr := os.Stat(path); statErr != nil {
		t.Errorf("expected db file at %s, got %v", path, statErr)
	}
}

func TestNewSQLite_FileURIWithQueryParams(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested")
	path := "file:" + filepath.Join(dir, "app.db") + "?_journal_mode=WAL"

	db, err := sqlite.NewSQLite(&sqlite.Config{Path: path})
	if err != nil {
		t.Fatalf("NewSQLite file: URI with query params: %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	if _, statErr := os.Stat(dir); statErr != nil {
		t.Errorf("expected parent dir %s to be created, got %v", dir, statErr)
	}
}

func TestNewSQLite_FileURIInMemory(t *testing.T) {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: "file::memory:?cache=shared"})
	if err != nil {
		t.Fatalf("NewSQLite file::memory: URI: %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	if err := db.Ping(); err != nil {
		t.Errorf("Ping file::memory: db: %v", err)
	}
}

func TestInitDefault_Success(t *testing.T) {
	prev := dbsqlx.Default()
	t.Cleanup(func() { dbsqlx.SetDefault(prev) })

	if err := sqlite.InitDefault(&sqlite.Config{Path: ":memory:"}); err != nil {
		t.Fatalf("InitDefault: %v", err)
	}
	if dbsqlx.Default() == nil {
		t.Error("Default() is nil after successful InitDefault")
	}
}

func TestInitDefault_Error(t *testing.T) {
	err := sqlite.InitDefault(&sqlite.Config{})
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("InitDefault empty config: err = %v, want ErrMissingConfig", err)
	}
}

func TestAttach(t *testing.T) {
	prev := dbsqlx.Default()
	t.Cleanup(func() { dbsqlx.SetDefault(prev) })

	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	t.Cleanup(func() { dbsqlx.Close(db) }) //nolint:errcheck
	dbsqlx.SetDefault(db)

	if err := sqlite.Attach(":memory:", "secondary"); err != nil {
		t.Errorf("Attach: %v", err)
	}
}
