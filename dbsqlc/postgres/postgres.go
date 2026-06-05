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

package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/phcp-tech/common-library-golang/dbsqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

// slog.Info / slog.Infof are used intentionally instead of the project's log package.
// slog uses the stdlib default logger, which works without any initialisation.
// When the caller invokes log.InitLog(), that function calls slog.SetDefault(l.Logger),
// so all subsequent slog calls are automatically routed to the project logger — no
// behaviour change is needed here.

type Config struct {
	// Database connection parameters
	Host     string
	Port     string
	Database string
	Username string
	Password string

	// Connection pool parameters
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdletime int

	// Optional search_path to set for all connections in the pool
	SearchPath string
}

// NewPostgres opens a *pgxpool.Pool connection to PostgreSQL and configures the connection pool.
// It returns the pgx native pool so sqlc-generated Queries (sql_package: "pgx/v5") can use it directly.
func NewPostgres(conf *Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.Database)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// If SearchPath is provided, set it as a runtime parameter for all connections in the pool.
	if conf.SearchPath != "" {
		poolConfig.ConnConfig.RuntimeParams["search_path"] = conf.SearchPath
	}

	// Set default values for connection pool settings
	poolConfig.MaxConns = int32(dbsqlc.MaxOpenConns)
	poolConfig.MinConns = int32(dbsqlc.MaxIdleConns)
	poolConfig.MaxConnLifetime = dbsqlc.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = dbsqlc.ConnMaxIdletime

	// Set connection pool settings from environment variables
	if conf.MaxOpenConns > 0 {
		poolConfig.MaxConns = int32(conf.MaxOpenConns)
	}
	if conf.MaxIdleConns > 0 {
		poolConfig.MinConns = int32(conf.MaxIdleConns)
	}
	if conf.ConnMaxLifetime > 0 {
		poolConfig.MaxConnLifetime = time.Duration(conf.ConnMaxLifetime) * time.Minute
	}
	if conf.ConnMaxIdletime > 0 {
		poolConfig.MaxConnIdleTime = time.Duration(conf.ConnMaxIdletime) * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	// Log runtime pool configuration
	slog.Info("Database pool configured", "maxConns", poolConfig.MaxConns, "minConns", poolConfig.MinConns)
	s := pool.Stat()
	slog.Info("Database stats",
		"totalConns", s.TotalConns(),
		"idleConns", s.IdleConns(),
		"acquiredConns", s.AcquiredConns(),
		"maxConns", s.MaxConns(),
	)

	// Verify search_path is set correctly on initial connection
	var path string
	if err := pool.QueryRow(context.Background(), "show search_path;").Scan(&path); err == nil {
		slog.Info("Show search_path", "path", path)
	} else {
		slog.Error("Failed to show search_path", "error", err)
	}

	return pool, nil
}
