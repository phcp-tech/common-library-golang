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
	Host       string
	Port       string
	Database   string
	SearchPath string // optional
	Username   string
	Password   string

	// Connection pool parameters
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdletime int
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
	slog.Info(fmt.Sprintf("Database pool configured - MaxConns: %d, MinConns: %d", poolConfig.MaxConns, poolConfig.MinConns))
	s := pool.Stat()
	slog.Info(fmt.Sprintf("Database stats - TotalConns: %d, IdleConns: %d, AcquiredConns: %d, MaxConns: %d", s.TotalConns(), s.IdleConns(), s.AcquiredConns(), s.MaxConns()))

	// Verify connectivity by issuing a lightweight query.
	// This converts pgxpool's lazy-connect behaviour into an eager check:
	// if the database is unreachable the caller receives an error immediately
	// rather than discovering it on the first real query.
	var path string
	if err := pool.QueryRow(context.Background(), "show search_path;").Scan(&path); err != nil {
		pool.Close()
		return nil, fmt.Errorf("show search_path failed: %w", err)
	}
	slog.Info(fmt.Sprintf("Show search_path: %s", path))

	return pool, nil
}
