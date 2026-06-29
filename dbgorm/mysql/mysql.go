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
	"log/slog"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config contains MySQL connection settings.
type Config struct {
	// Required connection settings.
	Host     string
	Port     string
	Database string
	Username string
	Password string

	// gorm.Config fields for connection pool settings.
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdletime int
	Logger          *slog.Logger
}

// Dialector returns a MySQL GORM dialector from conf.
func Dialector(conf *Config) (gorm.Dialector, error) {
	if conf.Host == "" || conf.Port == "" || conf.Database == "" || conf.Username == "" {
		return nil, dbgorm.ErrMissingConfig
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.Database)
	return gormmysql.Open(dsn), nil
}

// NewMySQL opens a MySQL-backed GORM database.
func NewMySQL(conf *Config) (*gorm.DB, error) {
	dialector, err := Dialector(conf)
	if err != nil {
		return nil, err
	}
	return dbgorm.Open(dialector, &dbgorm.GormConfig{
		MaxOpenConns:    conf.MaxOpenConns,
		MaxIdleConns:    conf.MaxIdleConns,
		ConnMaxLifetime: conf.ConnMaxLifetime,
		ConnMaxIdletime: conf.ConnMaxIdletime,
		Logger:          conf.Logger,
	})
}

// InitDefault opens MySQL and stores it as the dbgorm default database.
func InitDefault(conf *Config) error {
	db, err := NewMySQL(conf)
	if err != nil {
		return err
	}
	dbgorm.SetDefault(db)
	return nil
}
