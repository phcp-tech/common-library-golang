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

package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/phcp-tech/common-library-golang/dbsqlx"

	"github.com/vinovest/sqlx"
	_ "modernc.org/sqlite" // pure-Go driver (no CGO), registers as "sqlite"
)

// sqliteMaxOpenConns/sqliteMaxIdleConns are fixed rather than configurable:
// SQLite allows only one writer at a time, so a larger pool provides no
// benefit even with WAL mode enabled.
const (
	sqliteMaxOpenConns = 4
	sqliteMaxIdleConns = 2
)

// Config contains SQLite connection settings.
type Config struct {
	// Path is the path to the SQLite database file, or ":memory:" for an
	// in-memory database. URI format is also supported, e.g.
	// "file:app.db?_journal_mode=WAL&_foreign_keys=on".
	Path string
}

// pragmas are applied to every new SQLite connection after opening.
// WAL mode enables concurrent reads alongside a single writer.
var pragmas = []string{
	"PRAGMA journal_mode=WAL",   // concurrent read + write; one writer at a time
	"PRAGMA synchronous=NORMAL", // safe under WAL, faster than FULL
	"PRAGMA foreign_keys=ON",    // enforce FK constraints (SQLite default is OFF)
	"PRAGMA busy_timeout=5000",  // wait up to 5 s for write lock instead of SQLITE_BUSY
	"PRAGMA cache_size=-32000",  // 32 MB in-memory page cache
}

// ensureSQLiteDir creates the parent directory of a file-based SQLite path.
// It is a no-op for in-memory paths (":memory:" and "file::memory:...").
// SQLite can auto-create the database file but cannot create missing directories.
func ensureSQLiteDir(path string) error {
	if path == "" || path == ":memory:" {
		return nil
	}

	var fp string
	if strings.HasPrefix(path, "file:") {
		// strip "file:" prefix and drop query parameters
		s := strings.SplitN(path[5:], "?", 2)[0]
		if s != "" && !strings.HasPrefix(s, ":memory") {
			fp = s
		}
	} else {
		fp = path
	}

	if fp == "" {
		return nil
	}

	dir := filepath.Dir(fp)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("sqlite: create db dir: %w", err)
		}
	}
	return nil
}

// NewSQLite opens a SQLite-backed *sqlx.DB and applies the standard PRAGMA
// settings (WAL, foreign keys, busy timeout, cache size).
// The parent directory of a file-based path is created automatically if absent.
func NewSQLite(conf *Config) (*sqlx.DB, error) {
	if conf == nil || conf.Path == "" {
		return nil, dbsqlx.ErrMissingConfig
	}

	if err := ensureSQLiteDir(conf.Path); err != nil {
		return nil, err
	}

	// WAL mode is set below, which allows concurrent reads.
	// MaxOpenConns=4 is appropriate once WAL is active.
	db, err := dbsqlx.Open("sqlite", conf.Path, &dbsqlx.PoolConfig{
		MaxOpenConns: sqliteMaxOpenConns,
		MaxIdleConns: sqliteMaxIdleConns,
	})
	if err != nil {
		return nil, err
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return nil, fmt.Errorf("sqlite pragma failed %q: %w", p, err)
		}
	}

	return db, nil
}

// InitDefault opens SQLite and stores it as the dbsqlx default database.
func InitDefault(conf *Config) error {
	db, err := NewSQLite(conf)
	if err != nil {
		return err
	}
	dbsqlx.SetDefault(db)
	return nil
}

// Attach attaches an additional SQLite database file (dbfile) to the default
// database connection under the schema alias dbname using the SQLite ATTACH
// DATABASE statement.
func Attach(dbfile string, dbname string) error {
	_, err := dbsqlx.Default().Exec("ATTACH DATABASE '" + dbfile + "' as " + dbname)
	return err
}
