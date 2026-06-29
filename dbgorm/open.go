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

package dbgorm

import (
	"errors"
	"log/slog"
	"time"

	slogGorm "github.com/orandin/slog-gorm"
	"github.com/phcp-tech/common-library-golang/log"
	"gorm.io/gorm"
)

var (
	// ErrNilDialector is returned when Open receives a nil GORM dialector.
	ErrNilDialector = errors.New("dbgorm: nil gorm dialector")
	// ErrMissingConfig is returned when an adapter lacks required connection fields.
	ErrMissingConfig = errors.New("dbgorm: missing connection fields")
)

// IsNotFound reports whether err represents GORM's record-not-found condition.
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// GormConfig contains pool settings and logger for a GORM database connection.
type GormConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdletime int
	Logger          *slog.Logger
}

// Open opens a GORM database using the supplied dialector and common config.
func Open(dialector gorm.Dialector, conf *GormConfig) (*gorm.DB, error) {
	if dialector == nil {
		return nil, ErrNilDialector
	}

	// set up slog handler, using provided logger if available, otherwise use slog.Default()
	handler := slog.Default().Handler()
	if conf.Logger != nil {
		handler = conf.Logger.Handler()
	}

	// Create an slog-gorm instance
	gormLogger := slogGorm.New(
		slogGorm.WithHandler(handler), // Optional, use slog.Default() by default
		//slogGorm.WithTraceAll(),                               // trace all messages
	)

	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // Disable creation of associated foreign key constraints
		SkipDefaultTransaction:                   true, // Skip create, update, delete transactions for performance
		PrepareStmt:                              true, // Prepare statements for performance. When executing any SQL, a prepared statement will be created and cached to improve subsequent efficiency
		Logger:                                   gormLogger,
	}

	// ClickHouse HTTP/TCPNative protocol maps Prepare() to batch INSERT; PrepareStmt must be disabled for SELECT queries
	if dialector.Name() == "clickhouse" {
		gormConfig.PrepareStmt = false
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, err
	}

	// Set connection pool default settings
	maxOpenConns := MaxOpenConns
	maxIdleConns := MaxIdleConns
	connMaxLifetime := ConnMaxLifetime
	connMaxIdletime := ConnMaxIdletime

	// Override defaults with config values if provided
	if conf.MaxOpenConns > 0 {
		maxOpenConns = conf.MaxOpenConns
	}
	if conf.MaxIdleConns > 0 {
		maxIdleConns = conf.MaxIdleConns
	}
	if conf.ConnMaxLifetime > 0 {
		connMaxLifetime = time.Duration(conf.ConnMaxLifetime) * time.Minute
	}
	if conf.ConnMaxIdletime > 0 {
		connMaxIdletime = time.Duration(conf.ConnMaxIdletime) * time.Minute
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdletime)

	// Log runtime sql.DB configuration and immediate stats for debugging
	stats := sqlDB.Stats()
	log.Infof("Database pool configured: maxOpenConns=%d, maxIdleConns=%d, connMaxLifetime=%s, connMaxIdletime=%s",
		maxOpenConns, maxIdleConns, connMaxLifetime, connMaxIdletime)
	log.Infof("Database pool stats: OpenConnections=%d Idle=%d InUse=%d MaxOpenConnections=%d WaitCount=%d",
		stats.OpenConnections, stats.Idle, stats.InUse, stats.MaxOpenConnections, stats.WaitCount)

	return db, nil
}

// Close closes the underlying database/sql connection for db.
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, _ := db.DB()
	return sqlDB.Close()
}
