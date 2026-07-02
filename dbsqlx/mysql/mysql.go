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
	"fmt"

	"github.com/phcp-tech/common-library-golang/dbsqlx"

	_ "github.com/go-sql-driver/mysql" // registers "mysql" driver with database/sql
	"github.com/vinovest/sqlx"
)

// Config contains MySQL connection settings.
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

// DSN builds a go-sql-driver/mysql connection string from conf.
func DSN(conf *Config) (string, error) {
	if conf.Host == "" || conf.Port == "" || conf.Database == "" || conf.Username == "" {
		return "", dbsqlx.ErrMissingConfig
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.Database), nil
}

// NewMySQL opens a MySQL-backed *sqlx.DB via the go-sql-driver/mysql driver.
// Open verifies connectivity with an eager ping — a non-nil error is returned
// immediately when the database is unreachable.
func NewMySQL(conf *Config) (*sqlx.DB, error) {
	dsn, err := DSN(conf)
	if err != nil {
		return nil, err
	}

	return dbsqlx.Open("mysql", dsn, &dbsqlx.PoolConfig{
		MaxOpenConns:    conf.MaxOpenConns,
		MaxIdleConns:    conf.MaxIdleConns,
		ConnMaxLifetime: conf.ConnMaxLifetime,
		ConnMaxIdletime: conf.ConnMaxIdletime,
	})
}

// InitDefault opens MySQL and stores it as the dbsqlx default database.
func InitDefault(conf *Config) error {
	db, err := NewMySQL(conf)
	if err != nil {
		return err
	}
	dbsqlx.SetDefault(db)
	return nil
}
