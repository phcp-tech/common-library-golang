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

// Package component provides MySQL lifecycle integration for bootstrap.
package component

import (
	"log/slog"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/dbsqlc/mysql"
	"github.com/phcp-tech/common-library-golang/env"
)

// loadFromEnv reads MySQL connection parameters from the koanf env singleton
// and initialises the package-level default connection.
// Configuration keys:
//
//	db.host, db.port, db.name, db.username, db.password
//	db.max.open.conns, db.max.idle.conns
//	db.conn.max.lifetime, db.conn.max.idletime
//
// Note: mysql.InitDefault calls sql.Open which is lazy — no connection is
// established here. The first actual query or an explicit PingContext will
// attempt the network connection.
func loadFromEnv() error {
	config := &mysql.Config{
		Host:            env.Env().String("db.host"),
		Port:            env.Env().String("db.port"),
		Database:        env.Env().String("db.name"),
		Username:        env.Env().String("db.username"),
		Password:        env.Env().String("db.password"),
		MaxOpenConns:    env.Env().Int("db.max.open.conns"),
		MaxIdleConns:    env.Env().Int("db.max.idle.conns"),
		ConnMaxLifetime: env.Env().Int("db.conn.max.lifetime"),
		ConnMaxIdletime: env.Env().Int("db.conn.max.idletime"),
	}

	if err := mysql.InitDefault(config); err != nil {
		slog.Error("Connect database failed", "error", err)
		return err
	}
	slog.Info("Connect database successfully")
	return nil
}

// Component wraps loadFromEnv as a bootstrap.IComponent.
// Init calls loadFromEnv; Close closes the default MySQL connection.
//
// Because sql.Open is lazy, Init() always returns nil — the actual network
// connection is deferred to the first query. Use a PreReady step to ping the
// database if an eager connectivity check is required:
//
//	AddParallel(component.Component()).
//	PreReady(func() error {
//	    return mysql.Default().PingContext(ctx)
//	}).
//	Run()
func Component() bootstrap.IComponent {
	return bootstrap.Func(
		"mysql",
		loadFromEnv,
		func() {
			if db := mysql.Default(); db != nil {
				db.Close()
				slog.Info("Database has been closed")
			}
		},
	)
}
