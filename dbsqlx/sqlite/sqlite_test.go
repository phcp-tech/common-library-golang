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
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
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

	var busyTimeout int
	if err := db.Get(&busyTimeout, "PRAGMA busy_timeout"); err != nil {
		t.Fatalf("query busy_timeout: %v", err)
	}
	if busyTimeout != 5000 {
		t.Errorf("busy_timeout = %d, want 5000", busyTimeout)
	}

	var synchronous int
	if err := db.Get(&synchronous, "PRAGMA synchronous"); err != nil {
		t.Fatalf("query synchronous: %v", err)
	}
	if synchronous != 1 { // 1 = NORMAL
		t.Errorf("synchronous = %d, want 1 (NORMAL)", synchronous)
	}

	var cacheSize int
	if err := db.Get(&cacheSize, "PRAGMA cache_size"); err != nil {
		t.Fatalf("query cache_size: %v", err)
	}
	if cacheSize != -32000 {
		t.Errorf("cache_size = %d, want -32000", cacheSize)
	}
}

// TestNewSQLite_PragmasAppliedToEveryPooledConnection is a regression test:
// PRAGMAs used to be applied via a one-time db.Exec after opening, which
// only affected whichever single connection happened to service that Exec -
// every other connection database/sql later opened under MaxOpenConns>1
// silently kept SQLite's defaults (busy_timeout=0, foreign_keys=off, ...).
// db.Conn(ctx) forces the pool to hand out distinct physical connections
// (each one held open occupies its own pool slot instead of being reused),
// so this checks busy_timeout on more than just the first connection ever
// opened.
func TestNewSQLite_PragmasAppliedToEveryPooledConnection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pooled.db")
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: path})
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	ctx := context.Background()
	const n = 3 // < the pool's MaxOpenConns, so all n are guaranteed distinct
	conns := make([]*sql.Conn, n)
	for i := range conns {
		c, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("Conn() #%d: %v", i, err)
		}
		conns[i] = c
	}
	defer func() {
		for _, c := range conns {
			c.Close() //nolint:errcheck
		}
	}()

	for i, c := range conns {
		var busyTimeout int
		if err := c.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
			t.Fatalf("conn #%d: query busy_timeout: %v", i, err)
		}
		if busyTimeout != 5000 {
			t.Errorf("conn #%d: busy_timeout = %d, want 5000", i, busyTimeout)
		}
	}
}

// TestNewSQLite_ConcurrentWritesAcrossPoolDoNotFailWithBusy is a regression
// test for the same underlying bug as
// TestNewSQLite_PragmasAppliedToEveryPooledConnection, exercised end-to-end:
// two connections racing to write used to produce a genuine
// "database is locked (5) (SQLITE_BUSY)" instead of one waiting for the
// other's lock to release, because only one connection in the pool ever
// actually had a non-zero busy_timeout.
func TestNewSQLite_ConcurrentWritesAcrossPoolDoNotFailWithBusy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "concurrent.db")
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: path})
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, v TEXT)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	const goroutines = 20
	const writesEach = 20
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines*writesEach)
	for g := range goroutines {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := range writesEach {
				if _, err := db.Exec(`INSERT INTO t (v) VALUES (?)`, fmt.Sprintf("g%d-%d", g, i)); err != nil {
					errCh <- err
				}
			}
		}(g)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("write error (want none): %v", err)
	}

	var count int
	if err := db.Get(&count, `SELECT COUNT(*) FROM t`); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != goroutines*writesEach {
		t.Errorf("row count = %d, want %d", count, goroutines*writesEach)
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
