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

package dbsqlx_test

import (
	"errors"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
)

func TestOpen_MissingDriverName(t *testing.T) {
	_, err := dbsqlx.Open("", ":memory:", nil)
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("Open with empty driverName: err = %v, want ErrMissingConfig", err)
	}
}

func TestOpen_MissingDSN(t *testing.T) {
	_, err := dbsqlx.Open("sqlite", "", nil)
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("Open with empty dsn: err = %v, want ErrMissingConfig", err)
	}
}

func TestOpen_UnknownDriver(t *testing.T) {
	_, err := dbsqlx.Open("no-such-driver", ":memory:", nil)
	if err == nil {
		t.Error("Open with unregistered driver: want error, got nil")
	}
}

func TestOpen_Success(t *testing.T) {
	db, err := dbsqlx.Open("sqlite", ":memory:", nil)
	if err != nil {
		t.Fatalf("Open error = %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	if db == nil {
		t.Fatal("Open returned nil db with nil error")
	}
	if err := db.Ping(); err != nil {
		t.Errorf("Ping after Open: %v", err)
	}
}

func TestOpen_AppliesDefaultPoolSettings(t *testing.T) {
	db, err := dbsqlx.Open("sqlite", ":memory:", nil)
	if err != nil {
		t.Fatalf("Open error = %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	stats := db.Stats()
	if stats.MaxOpenConnections != dbsqlx.MaxOpenConns {
		t.Errorf("MaxOpenConnections = %d, want %d", stats.MaxOpenConnections, dbsqlx.MaxOpenConns)
	}
}

func TestOpen_AppliesCustomPoolSettings(t *testing.T) {
	db, err := dbsqlx.Open("sqlite", ":memory:", &dbsqlx.PoolConfig{
		MaxOpenConns:    7,
		MaxIdleConns:    3,
		ConnMaxLifetime: 30,
		ConnMaxIdletime: 5,
	})
	if err != nil {
		t.Fatalf("Open error = %v", err)
	}
	defer dbsqlx.Close(db) //nolint:errcheck

	stats := db.Stats()
	if stats.MaxOpenConnections != 7 {
		t.Errorf("MaxOpenConnections = %d, want 7", stats.MaxOpenConnections)
	}
}

func TestClose_NilDB(t *testing.T) {
	if err := dbsqlx.Close(nil); err != nil {
		t.Errorf("Close(nil) = %v, want nil", err)
	}
}

func TestClose_Success(t *testing.T) {
	db, err := dbsqlx.Open("sqlite", ":memory:", nil)
	if err != nil {
		t.Fatalf("Open error = %v", err)
	}
	if err := dbsqlx.Close(db); err != nil {
		t.Errorf("Close error = %v", err)
	}
}
