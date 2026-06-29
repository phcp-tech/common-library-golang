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

// Package component provides ClickHouse lifecycle integration for bootstrap.
package component

import (
	"log/slog"

	"github.com/phcp-tech/common-library-golang/bootstrap"
	"github.com/phcp-tech/common-library-golang/dbsqlc/clickhouse"
	"github.com/phcp-tech/common-library-golang/env"
)

// clickhouseInitDefault is the InitDefault implementation used by loadFromEnv.
// It is a package-level variable so that tests can replace it with a stub.
var clickhouseInitDefault = clickhouse.InitDefault

// loadFromEnv reads ClickHouse connection parameters from the koanf env singleton
// and initialises the package-level default client.
// Configuration keys:
//
//	db.host, db.port, db.name, db.username, db.password
//	db.max.open.conns, db.max.idle.conns, db.conn.max.lifetime
//
// Note: clickhouse-go opens connections lazily — Init() returns nil even when
// the server is unreachable. The first real operation (Ping/Query) will fail.
func loadFromEnv() error {
	config := &clickhouse.Config{
		Host:            env.Env().String("db.host"),
		Port:            env.Env().String("db.port"),
		Database:        env.Env().String("db.name"),
		Username:        env.Env().String("db.username"),
		Password:        env.Env().String("db.password"),
		MaxOpenConns:    env.Env().Int("db.max.open.conns"),
		MaxIdleConns:    env.Env().Int("db.max.idle.conns"),
		ConnMaxLifetime: env.Env().Int("db.conn.max.lifetime"),
	}

	if err := clickhouseInitDefault(config); err != nil {
		slog.Error("Connect ClickHouse failed", "error", err)
		return err
	}
	slog.Info("Connect ClickHouse successfully")
	return nil
}

// Component wraps loadFromEnv as a bootstrap.IComponent.
// Init calls loadFromEnv; Close closes the default ClickHouse connection.
//
// Because clickhouse-go opens connections lazily, Init() always returns nil.
// Use a PreReady step with conn.Ping() for an eager connectivity check:
//
//	Add(component.Component()).
//	PreReady(func() error {
//	    return clickhouse.Default().Ping(ctx)
//	}).
//	Run()
func Component() bootstrap.IComponent {
	return bootstrap.Func(
		"clickhouse",
		loadFromEnv,
		func() {
			conn := clickhouse.Default()
			if conn == nil {
				return
			}
			if err := conn.Close(); err != nil {
				slog.Error("ClickHouse close failed", "error", err)
				return
			}
			slog.Info("ClickHouse has been closed")
		},
	)
}
