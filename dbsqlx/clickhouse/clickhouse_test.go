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

package clickhouse_test

import (
	"errors"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dbsqlx/clickhouse"
)

func TestDSN_FromStructuredFields(t *testing.T) {
	dsn, err := clickhouse.DSN(&clickhouse.Config{
		Host:     "localhost",
		Port:     "9440",
		Database: "risk",
		Username: "risk",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("DSN: %v", err)
	}
	want := "clickhouse://risk:secret@localhost:9440/risk?secure=true&skip_verify=true"
	if dsn != want {
		t.Errorf("DSN = %q, want %q", dsn, want)
	}
}

func TestDSN_RequiresStructuredFields(t *testing.T) {
	if _, err := clickhouse.DSN(&clickhouse.Config{}); !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Fatalf("DSN empty config: err = %v, want ErrMissingConfig", err)
	}
}

func TestDSN_MissingEachRequiredField(t *testing.T) {
	base := clickhouse.Config{Host: "localhost", Port: "9440", Database: "risk", Username: "risk", Password: "secret"}

	tests := []struct {
		name   string
		mutate func(*clickhouse.Config)
	}{
		{"missing host", func(c *clickhouse.Config) { c.Host = "" }},
		{"missing port", func(c *clickhouse.Config) { c.Port = "" }},
		{"missing database", func(c *clickhouse.Config) { c.Database = "" }},
		{"missing username", func(c *clickhouse.Config) { c.Username = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := base
			tt.mutate(&conf)
			if _, err := clickhouse.DSN(&conf); !errors.Is(err, dbsqlx.ErrMissingConfig) {
				t.Errorf("DSN with %s: err = %v, want ErrMissingConfig", tt.name, err)
			}
		})
	}
}

// TestNewClickHouse_ErrMissingConfig covers the DSN-error path in NewClickHouse.
func TestNewClickHouse_ErrMissingConfig(t *testing.T) {
	_, err := clickhouse.NewClickHouse(&clickhouse.Config{}) // all required fields empty
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("NewClickHouse empty config: want ErrMissingConfig, got %v", err)
	}
}

// TestInitDefault_Error covers the error-return path in InitDefault.
func TestInitDefault_Error(t *testing.T) {
	err := clickhouse.InitDefault(&clickhouse.Config{}) // DSN fails → error propagated
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("InitDefault empty config: want ErrMissingConfig, got %v", err)
	}
}
