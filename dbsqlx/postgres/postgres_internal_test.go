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

// Package postgres (internal test) covers private helpers unreachable from
// the external test package without a live PostgreSQL server.
package postgres

import (
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dbsqlx/sqlite"
)

// TestLogSearchPath_WarnPath exercises the slog.Warn branch of logSearchPath.
// SQLite does not support "SHOW search_path", so the query fails and Warn is logged.
func TestLogSearchPath_WarnPath(t *testing.T) {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("sqlite.NewSQLite: %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	logSearchPath(db, "SHOW search_path") // SQLite rejects → Warn path
}

// TestLogSearchPath_InfoPath exercises the slog.Info branch of logSearchPath.
// A query that SQLite can execute returns nil error, so the Info branch is taken
// without requiring a live PostgreSQL connection.
func TestLogSearchPath_InfoPath(t *testing.T) {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("sqlite.NewSQLite: %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	logSearchPath(db, "SELECT 'public'") // SQLite succeeds → Info path
}
