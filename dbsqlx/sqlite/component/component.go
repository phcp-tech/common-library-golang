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

// Package component provides sqlx SQLite lifecycle integration for bootstrap.
package component

import (
	"log/slog"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dbsqlx/sqlite"
	"github.com/phcp-tech/common-library-golang/env"
)

// sqliteInitDefault is the InitDefault implementation used by loadFromEnv.
// It is a package-level variable so that tests can replace it with a stub.
var sqliteInitDefault = sqlite.InitDefault

// loadFromEnv reads the SQLite database path from the koanf env singleton
// and initialises the process-wide default sqlx database.
// Configuration key:
//
//	db.sqlite.path — file path or ":memory:" for in-memory databases
//	                 e.g. "file:app.db?cache=shared"
//
// SQLite is an embedded database — no network connection is attempted.
// Init() fails only when the path is empty or the file cannot be created.
func loadFromEnv() error {
	config := &sqlite.Config{
		Path: env.Env().String("db.sqlite.path"),
	}

	if err := sqliteInitDefault(config); err != nil {
		slog.Error("Connect SQLite failed", "error", err)
		return err
	}
	slog.Info("Connect SQLite successfully")
	return nil
}

// Component wraps loadFromEnv as a bootstrap.IComponent.
// Init calls loadFromEnv; Close closes the default sqlx SQLite connection.
//
// The SQLite path is read from env key db.sqlite.path during Init().
// Typical values:
//   - "file:app.db?cache=shared" for file-based databases
//   - "file::memory:?cache=shared" for in-memory databases (tests)
func Component() bootstrap.IComponent {
	return bootstrap.Func(
		"sqlite",
		loadFromEnv,
		func() {
			db := dbsqlx.Default()
			if db == nil {
				return
			}
			if err := dbsqlx.Close(db); err != nil {
				slog.Error("SQLite close failed", "error", err)
				return
			}
			slog.Info("SQLite has been closed")
		},
	)
}
