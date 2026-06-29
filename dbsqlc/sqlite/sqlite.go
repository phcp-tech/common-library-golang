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
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	sqliteMaxOpenConns = 4
	sqliteMaxIdleConns = 2
)

// Config holds SQLite connection parameters.
type Config struct {
	// Path is the path to the SQLite database file, or ":memory:" for an in-memory database.
	// URI format is supported: "file:app.db?_journal_mode=WAL&_foreign_keys=on"
	Path string
}

// NewSQLite opens a *sql.DB connection to SQLite.
// modernc.org/sqlite is a pure-Go driver (no CGO required).
//
// For file-based databases, WAL mode and foreign key enforcement are recommended:
//
//	conf := &Config{Path: "file:app.db?_journal_mode=WAL&_foreign_keys=on"}
func NewSQLite(conf *Config) (*sql.DB, error) {
	// ensure parent directory exists for file-based DB paths
	if conf != nil && conf.Path != "" {
		p := conf.Path
		// skip in-memory variants
		if p != ":memory:" {
			var fp string
			if strings.HasPrefix(p, "file:") {
				s := strings.SplitN(p[5:], "?", 2)[0]
				if s != "" && !strings.HasPrefix(s, ":memory") {
					fp = s
				}
			} else {
				fp = p
			}

			// sqlite can create db file automaticlly but can't create folder automaticlly
			if fp != "" {
				dir := filepath.Dir(fp)
				if dir != "." && dir != "" {
					if err := os.MkdirAll(dir, 0o755); err != nil {
						return nil, fmt.Errorf("create db dir: %w", err)
					}
				}
			}
		}
	}

	db, err := sql.Open("sqlite", conf.Path)
	if err != nil {
		return nil, err
	}

	// key configs
	pragmas := []string{
		"PRAGMA journal_mode=WAL",   // allow concurrent read and write, significantly improving performance
		"PRAGMA synchronous=NORMAL", // secure and faster under WAL mode
		"PRAGMA foreign_keys=ON",    // enforce foreign key constraints
		"PRAGMA busy_timeout=5000",  // wait 5s for write lock instead of immediate error
		"PRAGMA cache_size=-32000",  // 32MB memory cache
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return nil, fmt.Errorf("pragma failed %s: %w", p, err)
		}
	}

	// SQLite allows only one writer at a time;
	// without WAL mode, call SetMaxOpenConns(sqliteMaxOpenConns) on the returned *sql.DB to prevent "database is locked" errors.
	// with WAL mode, multiple connections can read and write concurrently, so we can set a reasonable pool size.
	db.SetMaxOpenConns(sqliteMaxOpenConns)
	db.SetMaxIdleConns(sqliteMaxIdleConns)

	return db, nil
}
