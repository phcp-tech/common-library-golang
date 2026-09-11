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

// defaultPragmaParams is appended to conf.Path (see withDefaultPragmas) so
// every physical connection modernc.org/sqlite opens for this DSN gets these
// PRAGMAs applied - not just whichever single connection happens to service
// a one-time db.Exec("PRAGMA ...") call after NewSQLite returns, which is
// what this used to do. That distinction matters because journal_mode is a
// persistent property of the database file itself (once one connection sets
// WAL, every connection sees WAL, DSN or not) but busy_timeout/synchronous/
// foreign_keys/cache_size are per-connection runtime state that SQLite never
// persists - with MaxOpenConns > 1, only the one connection the old Exec
// loop happened to land on ever actually got a non-zero busy_timeout. Every
// other pooled connection defaulted to busy_timeout=0 - "fail immediately"
// instead of "wait for the lock" - which surfaced as a real, user-visible
// "database is locked (5) (SQLITE_BUSY)" the moment two connections tried to
// write around the same time (e.g. a request's own DAO write racing an
// async background write on a different pooled connection). DSN query
// parameters, by contrast, are re-applied by the driver on every Open call
// (see modernc.org/sqlite's driver.go doc comment for the _busy_timeout/
// _journal_mode/_synchronous/_foreign_keys shorthand keys - cache_size has
// no shorthand, hence the raw _pragma= form for it), so this fix covers
// every connection the pool ever opens, not just the first one.
const defaultPragmaParams = "_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=on&_pragma=cache_size(-32000)"

// withDefaultPragmas appends defaultPragmaParams to path - see its doc
// comment for why this must happen per-connection via the DSN rather than
// once via db.Exec. If path already carries its own query string (the
// documented "URI format... file:app.db?_journal_mode=WAL&_foreign_keys=on"
// escape hatch), ours are appended after with "&": modernc.org/sqlite
// resolves a repeated key to whichever occurrence comes first (Go's
// url.Values.Get semantics), so a caller's own explicit value for any of
// these keys wins over our default instead of being silently overridden.
func withDefaultPragmas(path string) string {
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + defaultPragmaParams
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
// settings (WAL, foreign keys, busy timeout, cache size) to every connection
// it opens, now and for the lifetime of the pool - see withDefaultPragmas'
// doc comment for why these are carried on the DSN itself rather than run
// once via Exec after opening. The parent directory of a file-based path is
// created automatically if absent.
func NewSQLite(conf *Config) (*sqlx.DB, error) {
	if conf == nil || conf.Path == "" {
		return nil, dbsqlx.ErrMissingConfig
	}

	if err := ensureSQLiteDir(conf.Path); err != nil {
		return nil, err
	}

	// MaxOpenConns=4 is safe because every one of those connections gets
	// WAL mode and a real busy_timeout via the DSN (withDefaultPragmas),
	// not just whichever connection served a one-time PRAGMA Exec.
	db, err := dbsqlx.Open("sqlite", withDefaultPragmas(conf.Path), &dbsqlx.PoolConfig{
		MaxOpenConns: sqliteMaxOpenConns,
		MaxIdleConns: sqliteMaxIdleConns,
	})
	if err != nil {
		return nil, err
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
