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
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"

	gormsqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	sqliteMaxOpenConns = 4
	sqliteMaxIdleConns = 2
)

// Config contains SQLite connection settings.
type Config struct {
	Path string

	// gorm.Config fields for connection pool settings.
	Logger *slog.Logger
}

// Dialector returns a SQLite GORM dialector from conf.
func Dialector(conf *Config) (gorm.Dialector, error) {
	if conf.Path == "" {
		return nil, dbgorm.ErrMissingConfig
	}
	return gormsqlite.Open(conf.Path), nil
}

// pragmas are applied to every new SQLite connection after opening.
// WAL mode enables concurrent reads alongside a single writer.
// With WAL active, MaxOpenConns=4 is safe; without it concurrent writers
// would cause "database is locked" errors.
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

// NewSQLite opens a SQLite-backed GORM database and applies the standard
// PRAGMA settings (WAL, foreign keys, busy timeout, cache size).
// The parent directory of a file-based path is created automatically if absent.
func NewSQLite(conf *Config) (*gorm.DB, error) {
	if err := ensureSQLiteDir(conf.Path); err != nil {
		return nil, err
	}

	dialector, err := Dialector(conf)
	if err != nil {
		return nil, err
	}

	// WAL mode is set below, which allows concurrent reads.
	// MaxOpenConns=4 is appropriate once WAL is active.
	db, err := dbgorm.Open(dialector, &dbgorm.GormConfig{
		MaxOpenConns: sqliteMaxOpenConns,
		MaxIdleConns: sqliteMaxIdleConns,
		Logger:       conf.Logger,
	})
	if err != nil {
		return nil, err
	}

	for _, p := range pragmas {
		if err := db.Exec(p).Error; err != nil {
			return nil, fmt.Errorf("sqlite pragma failed %q: %w", p, err)
		}
	}

	return db, nil
}

// InitDefault opens SQLite and stores it as the dbgorm default database.
func InitDefault(conf *Config) error {
	db, err := NewSQLite(conf)
	if err != nil {
		return err
	}
	dbgorm.SetDefault(db)
	return nil
}

// Attach attaches an additional SQLite database file (dbfile) to the global database
// connection under the schema alias dbname using the SQLite ATTACH DATABASE statement.
// Returns the resulting *gorm.DB so errors can be checked via .Error.
func Attach(dbfile string, dbname string) *gorm.DB {
	return dbgorm.Default().Exec("ATTACH DATABASE '" + dbfile + "' as " + dbname)
}
