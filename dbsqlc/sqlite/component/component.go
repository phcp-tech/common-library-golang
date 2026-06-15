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

// Package component provides SQLite lifecycle integration for bootstrap.
package component

import (
	"log/slog"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/dbsqlc/sqlite"
	"github.com/phcp-tech/common-library-golang/env"
)

// loadFromEnv reads the SQLite database path from the koanf env singleton
// and initialises the package-level default connection.
// Configuration key:
//
//	db.sqlite.path — file path, ":memory:", or URI format
//	                 e.g. "file:app.db?_journal_mode=WAL&_foreign_keys=on"
func loadFromEnv() error {
	config := &sqlite.Config{
		Path: env.Env().String("db.sqlite.path"),
	}
	if err := sqlite.InitDefault(config); err != nil {
		slog.Error("Connect SQLite failed", "error", err)
		return err
	}
	slog.Info("Connect SQLite successfully")
	return nil
}

// Component wraps loadFromEnv as a bootstrap.IComponent.
// Init calls loadFromEnv; Close closes the default SQLite connection.
//
// The SQLite path is read from env key db.sqlite.path during Init().
// Typical values:
//   - ":memory:" for in-memory databases (tests, ephemeral data)
//   - "file:app.db?_journal_mode=WAL&_foreign_keys=on" for file-based databases
func Component() bootstrap.IComponent {
	return bootstrap.Func(
		"sqlite",
		loadFromEnv,
		func() {
			if db := sqlite.Default(); db != nil {
				db.Close()
				slog.Info("SQLite has been closed")
			}
		},
	)
}
