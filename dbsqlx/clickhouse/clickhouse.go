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
	"fmt"

	"github.com/phcp-tech/common-library-golang/dbsqlx"

	_ "github.com/ClickHouse/clickhouse-go/v2" // registers "clickhouse" driver with database/sql
	"github.com/vinovest/sqlx"
)

// Config contains ClickHouse connection settings.
type Config struct {
	// Required connection settings.
	Host     string
	Port     string
	Database string
	Username string
	Password string

	// Pool tuning; zero values fall back to dbsqlx package defaults.
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdletime int
}

// DSN builds a clickhouse-go/v2 connection string from conf.
func DSN(conf *Config) (string, error) {
	if conf.Host == "" || conf.Port == "" || conf.Database == "" || conf.Username == "" {
		return "", dbsqlx.ErrMissingConfig
	}

	// Native TCP + TLS (port 9440). Use "clickhouse://" scheme with secure=true for native binary protocol.
	// For HTTP/HTTPS (port 8123/8443) use "http://" or "https://" scheme instead.
	return fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s?secure=true&skip_verify=true",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.Database), nil
}

// NewClickHouse opens a ClickHouse-backed *sqlx.DB via the clickhouse-go/v2
// database/sql driver. Open verifies connectivity with an eager ping — a
// non-nil error is returned immediately when the database is unreachable.
func NewClickHouse(conf *Config) (*sqlx.DB, error) {
	dsn, err := DSN(conf)
	if err != nil {
		return nil, err
	}

	return dbsqlx.Open("clickhouse", dsn, &dbsqlx.PoolConfig{
		MaxOpenConns:    conf.MaxOpenConns,
		MaxIdleConns:    conf.MaxIdleConns,
		ConnMaxLifetime: conf.ConnMaxLifetime,
		ConnMaxIdletime: conf.ConnMaxIdletime,
	})
}

// InitDefault opens ClickHouse and stores it as the dbsqlx default database.
func InitDefault(conf *Config) error {
	db, err := NewClickHouse(conf)
	if err != nil {
		return err
	}
	dbsqlx.SetDefault(db)
	return nil
}
