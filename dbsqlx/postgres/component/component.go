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

// Package component provides sqlx PostgreSQL lifecycle integration for bootstrap.
package component

import (
	"log/slog"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dbsqlx/postgres"
	"github.com/phcp-tech/common-library-golang/env"
)

// postgresInitDefault is the InitDefault implementation used by loadFromEnv.
// It is a package-level variable so that tests can replace it with a stub.
var postgresInitDefault = postgres.InitDefault

// loadFromEnv reads PostgreSQL connection parameters from the koanf env singleton
// and initialises the process-wide default sqlx database.
// Configuration keys:
//
//	db.host, db.port, db.name, db.schema, db.username, db.password
//	db.max.open.conns, db.max.idle.conns
//	db.conn.max.lifetime, db.conn.max.idletime
//
// Note: Open verifies connectivity with an eager ping and runs SHOW search_path —
// Init() returns a non-nil error immediately when the database is unreachable.
func loadFromEnv() error {
	config := &postgres.Config{
		Host:            env.Env().String("db.host"),
		Port:            env.Env().String("db.port"),
		Database:        env.Env().String("db.name"),
		SearchPath:      env.Env().String("db.schema"),
		Username:        env.Env().String("db.username"),
		Password:        env.Env().String("db.password"),
		MaxOpenConns:    env.Env().Int("db.max.open.conns"),
		MaxIdleConns:    env.Env().Int("db.max.idle.conns"),
		ConnMaxLifetime: env.Env().Int("db.conn.max.lifetime"),
		ConnMaxIdletime: env.Env().Int("db.conn.max.idletime"),
	}

	if err := postgresInitDefault(config); err != nil {
		slog.Error("Connect database failed", "error", err)
		return err
	}
	slog.Info("Connect database successfully")
	return nil
}

// Component wraps loadFromEnv as a bootstrap.IComponent.
// Init calls loadFromEnv; Close closes the default sqlx database connection.
//
// Open verifies connectivity with an eager ping — if the database is
// unreachable, Init() returns a non-nil error and bootstrap aborts startup.
func Component() bootstrap.IComponent {
	return bootstrap.Func(
		"postgres",
		loadFromEnv,
		func() {
			db := dbsqlx.Default()
			if db == nil {
				return
			}
			if err := dbsqlx.Close(db); err != nil {
				slog.Error("Database close failed", "error", err)
				return
			}
			slog.Info("Database has been closed")
		},
	)
}
