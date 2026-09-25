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
	"fmt"
	"log/slog"
	"strings"

	"github.com/phcp-tech/common-library-golang/dbsqlx"

	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver with database/sql
	"github.com/vinovest/sqlx"
)

// Config contains PostgreSQL connection settings.
type Config struct {
	// Required connection settings.
	Host       string
	Port       string
	Database   string
	Username   string
	Password   string
	SearchPath string
	// QueryExecMode selects a pgx query execution mode. The zero value preserves
	// pgx's default; see the QueryExecMode constants for mode-specific constraints.
	QueryExecMode string

	// Pool tuning; zero values fall back to dbsqlx package defaults.
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdletime int
}

const (
	// QueryExecModeExec uses the extended protocol without a prepared-statement cache.
	QueryExecModeExec = "exec"
	// QueryExecModeCacheStatement enables pgx's prepared-statement cache.
	QueryExecModeCacheStatement = "cache_statement"
	// QueryExecModeCacheDescribe caches parameter and result type descriptions.
	QueryExecModeCacheDescribe = "cache_describe"
	// QueryExecModeDescribeExec describes a statement before every execution.
	QueryExecModeDescribeExec = "describe_exec"
	// QueryExecModeSimpleProtocol uses pgx's simple query protocol.
	QueryExecModeSimpleProtocol = "simple_protocol"
)

// DSN builds a libpq-style connection string from conf.
func DSN(conf *Config) (string, error) {
	if conf.Host == "" || conf.Port == "" || conf.Database == "" || conf.Username == "" {
		return "", dbsqlx.ErrMissingConfig
	}
	queryExecMode, err := normalizeQueryExecMode(conf.QueryExecMode)
	if err != nil {
		return "", err
	}

	// Another way for search_path: options='-c search_path=path1,path2'
	//
	// exec (the default) disables pgx's prepared-statement cache, which is
	// required by PgBouncer/Supavisor transaction-mode poolers such as
	// Supabase port 6543. It retains extended-protocol parameter type encoding,
	// including JSON/JSONB; cache_statement is opt-in for stable connections.
	parts := []string{
		fmt.Sprintf("host=%s", conf.Host),
		fmt.Sprintf("port=%s", conf.Port),
		fmt.Sprintf("user=%s", conf.Username),
		fmt.Sprintf("password=%s", conf.Password),
		fmt.Sprintf("dbname=%s", conf.Database),
		"sslmode=disable",
		"TimeZone=UTC",
	}
	if queryExecMode != "" {
		parts = append(parts, "default_query_exec_mode="+queryExecMode)
	}
	if conf.SearchPath != "" {
		parts = append(parts, fmt.Sprintf("search_path=%s", conf.SearchPath))
	}
	return strings.Join(parts, " "), nil
}

func normalizeQueryExecMode(mode string) (string, error) {
	switch mode {
	case "":
		return "", nil
	case QueryExecModeExec:
		return QueryExecModeExec, nil
	case QueryExecModeCacheStatement:
		return QueryExecModeCacheStatement, nil
	case QueryExecModeCacheDescribe:
		return QueryExecModeCacheDescribe, nil
	case QueryExecModeDescribeExec:
		return QueryExecModeDescribeExec, nil
	case QueryExecModeSimpleProtocol:
		return QueryExecModeSimpleProtocol, nil
	default:
		return "", fmt.Errorf("unsupported PostgreSQL query execution mode %q", mode)
	}
}

// NewPostgres opens a PostgreSQL-backed *sqlx.DB via the pgx stdlib driver.
// Open verifies connectivity with an eager ping — a non-nil error is returned
// immediately when the database is unreachable.
func NewPostgres(conf *Config) (*sqlx.DB, error) {
	dsn, err := DSN(conf)
	if err != nil {
		return nil, err
	}

	db, err := dbsqlx.Open("pgx", dsn, &dbsqlx.PoolConfig{
		MaxOpenConns:    conf.MaxOpenConns,
		MaxIdleConns:    conf.MaxIdleConns,
		ConnMaxLifetime: conf.ConnMaxLifetime,
		ConnMaxIdletime: conf.ConnMaxIdletime,
	})
	if err != nil {
		return nil, err
	}

	logSearchPath(db, "SHOW search_path")
	return db, nil
}

// logSearchPath executes query and logs the result as the current search_path.
// Errors are logged as warnings and do not abort startup.
// The query parameter is injectable so the function can be unit-tested without
// a live PostgreSQL server (production always passes "SHOW search_path").
func logSearchPath(db *sqlx.DB, query string) {
	var searchPath string
	if err := db.Get(&searchPath, query); err != nil {
		slog.Warn("PostgreSQL search_path check failed", "error", err)
	} else {
		slog.Info("PostgreSQL search_path", "search_path", searchPath)
	}
}

// InitDefault opens PostgreSQL and stores it as the dbsqlx default database.
func InitDefault(conf *Config) error {
	db, err := NewPostgres(conf)
	if err != nil {
		return err
	}
	dbsqlx.SetDefault(db)
	return nil
}
