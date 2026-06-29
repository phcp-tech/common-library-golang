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

package mysql

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/phcp-tech/common-library-golang/dbsqlc"

	_ "github.com/go-sql-driver/mysql"
)

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
	ConnMaxIdletime int
}

// NewMySQL opens a *sql.DB connection to MySQL and configures the connection pool.
// It returns the raw standard-library handle so sqlc-generated Queries can use it directly.
func NewMySQL(conf *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Set default values for connection pool settings
	db.SetMaxOpenConns(dbsqlc.MaxOpenConns)
	db.SetMaxIdleConns(dbsqlc.MaxIdleConns)
	db.SetConnMaxLifetime(dbsqlc.ConnMaxLifetime)
	db.SetConnMaxIdleTime(dbsqlc.ConnMaxIdletime)

	// Set connection pool settings from environment variables
	if conf.MaxOpenConns > 0 {
		db.SetMaxOpenConns(conf.MaxOpenConns)
	}
	if conf.MaxIdleConns > 0 {
		db.SetMaxIdleConns(conf.MaxIdleConns)
	}
	if conf.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Minute)
	}
	if conf.ConnMaxIdletime > 0 {
		db.SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdletime) * time.Minute)
	}

	// Log runtime pool configuration
	slog.Info("Database pool configured",
		"maxOpenConns", conf.MaxOpenConns,
		"maxIdleConns", conf.MaxIdleConns,
	)
	stats := db.Stats()
	slog.Info("Database stats",
		"openConnections", stats.OpenConnections,
		"idle", stats.Idle,
		"inUse", stats.InUse,
		"maxOpenConnections", stats.MaxOpenConnections,
		"waitCount", stats.WaitCount,
	)

	return db, nil
}
