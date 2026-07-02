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

// Package dbsqlx is a thin wrapper around github.com/vinovest/sqlx (a
// database/sql extension with reflection-based struct scanning and Go
// generics). It does not replace sqlx or hide *sqlx.DB — its purpose is to
// make connection setup consistent across drivers while keeping normal query
// execution in plain sqlx.
//
// The root package does not import concrete database drivers. Concrete
// drivers live in adapter packages (e.g. dbsqlx/sqlite), so applications only
// pull in the driver they actually use.
package dbsqlx

import (
	"errors"
	"log/slog"
	"time"

	"github.com/vinovest/sqlx"
)

// ErrMissingConfig is returned when an adapter lacks required connection fields.
var ErrMissingConfig = errors.New("dbsqlx: missing connection fields")

// Open connects to driverName using dsn and verifies connectivity with an
// eager ping, then applies pool settings from conf. Zero-value or nil conf
// fields fall back to the package defaults (MaxOpenConns, MaxIdleConns, etc).
func Open(driverName, dsn string, conf *PoolConfig) (*sqlx.DB, error) {
	if driverName == "" || dsn == "" {
		return nil, ErrMissingConfig
	}

	db, err := sqlx.Connect(driverName, dsn)
	if err != nil {
		return nil, err
	}

	maxOpenConns := MaxOpenConns
	maxIdleConns := MaxIdleConns
	connMaxLifetime := ConnMaxLifetime
	connMaxIdletime := ConnMaxIdletime

	if conf != nil {
		if conf.MaxOpenConns > 0 {
			maxOpenConns = conf.MaxOpenConns
		}
		if conf.MaxIdleConns > 0 {
			maxIdleConns = conf.MaxIdleConns
		}
		if conf.ConnMaxLifetime > 0 {
			connMaxLifetime = time.Duration(conf.ConnMaxLifetime) * time.Minute
		}
		if conf.ConnMaxIdletime > 0 {
			connMaxIdletime = time.Duration(conf.ConnMaxIdletime) * time.Minute
		}
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdletime)

	stats := db.Stats()
	slog.Info("Database pool configured",
		"driver", driverName,
		"maxOpenConns", maxOpenConns,
		"maxIdleConns", maxIdleConns,
		"connMaxLifetime", connMaxLifetime,
		"connMaxIdletime", connMaxIdletime,
	)
	slog.Info("Database pool stats",
		"openConnections", stats.OpenConnections,
		"idle", stats.Idle,
		"inUse", stats.InUse,
	)

	return db, nil
}

// Close closes the underlying database/sql connection for db.
func Close(db *sqlx.DB) error {
	if db == nil {
		return nil
	}
	return db.Close()
}
