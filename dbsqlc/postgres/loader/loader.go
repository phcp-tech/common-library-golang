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

package loader

import (
	"log/slog"

	"github.com/phcp-tech/common-library-golang/dbsqlc/postgres"
	"github.com/phcp-tech/common-library-golang/env"
)

func LoadFromEnv() error {
	config := &postgres.Config{
		// Load database connection parameters from environment variables
		Host:       env.Env().String("db.host"),
		Port:       env.Env().String("db.port"),
		Database:   env.Env().String("db.name"),
		SearchPath: env.Env().String("db.schema"),
		Username:   env.Env().String("db.username"),
		Password:   env.Env().String("db.password"),
		// Load connection pool settings from environment variables, with defaults if not set
		MaxOpenConns:    env.Env().Int("db.max.open.conns"),
		MaxIdleConns:    env.Env().Int("db.max.idle.conns"),
		ConnMaxLifetime: env.Env().Int("db.conn.max.lifetime"),
		ConnMaxIdletime: env.Env().Int("db.conn.max.idletime"),
	}

	if err := postgres.InitDefault(config); err != nil {
		slog.Info("Connect database failed: " + err.Error())
		return err
	}
	slog.Info("Connect database successfully")
	return nil
}
