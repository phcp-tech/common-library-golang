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

package clickhouse

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"

	"github.com/phcp-tech/common-library-golang/dbsqlc"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Package clickhouse provides a ClickHouse client using the native TCP protocol
// via github.com/ClickHouse/clickhouse-go/v2.
//
// Note: this package uses the ClickHouse native driver (driver.Conn), which is
// NOT compatible with sqlc code generation. sqlc requires a database/sql-compatible
// driver; for sqlc use the clickhouse-go HTTP interface with sql.Open("clickhouse", dsn)
// instead. This package is intended for direct, high-performance native protocol access.

// slog is used instead of the project log package so that this package has no
// initialisation dependency. slog uses the stdlib default logger when called
// standalone; when the caller invokes log.InitLog(), that function calls
// slog.SetDefault(l.Logger), so all slog calls are automatically routed to the
// project logger with no behaviour change here.

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
}

func NewClickHouse(conf *Config) (driver.Conn, error) {
	// Set default values for connection pool settings
	maxOpenConns := dbsqlc.MaxOpenConns
	maxIdleConns := dbsqlc.MaxIdleConns
	connMaxLifetime := dbsqlc.ConnMaxLifetime

	// Set connection pool settings from environment variables
	if conf.MaxOpenConns > 0 {
		maxOpenConns = conf.MaxOpenConns
	}
	if conf.MaxIdleConns > 0 {
		maxIdleConns = conf.MaxIdleConns
	}
	if conf.ConnMaxLifetime > 0 {
		connMaxLifetime = time.Duration(conf.ConnMaxLifetime) * time.Minute
	}

	db, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", conf.Host, conf.Port)},
		Auth: clickhouse.Auth{
			Database: conf.Database,
			Username: conf.Username,
			Password: conf.Password,
		},
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		// The clickhouse-go driver does not have a separate MaxIdletime setting
		//ConnMaxIdleTime: connMaxIdleTime,
	})

	if err != nil {
		return nil, err
	}

	// Log runtime pool configuration
	slog.Info("Database pool configured",
		"maxOpenConns", maxOpenConns,
		"maxIdleConns", maxIdleConns,
	)
	stats := db.Stats()
	slog.Info("Database stats",
		"open", stats.Open,
		"idle", stats.Idle,
		"maxOpenConns", stats.MaxOpenConns,
		"maxIdleConns", stats.MaxIdleConns,
	)

	return db, nil
}
