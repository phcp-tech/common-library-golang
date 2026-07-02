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

package mysql_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dbsqlx/mysql"
)

func TestDSN_FromStructuredFields(t *testing.T) {
	dsn, err := mysql.DSN(&mysql.Config{
		Host:     "localhost",
		Port:     "3306",
		Database: "risk",
		Username: "risk",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("DSN: %v", err)
	}
	want := "risk:secret@tcp(localhost:3306)/risk?charset=utf8mb4&parseTime=True&loc=UTC"
	if dsn != want {
		t.Errorf("DSN = %q, want %q", dsn, want)
	}
}

func TestDSN_RequiresStructuredFields(t *testing.T) {
	if _, err := mysql.DSN(&mysql.Config{}); !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Fatalf("DSN empty config: err = %v, want ErrMissingConfig", err)
	}
}

func TestDSN_MissingEachRequiredField(t *testing.T) {
	base := mysql.Config{Host: "localhost", Port: "3306", Database: "risk", Username: "risk", Password: "secret"}

	tests := []struct {
		name   string
		mutate func(*mysql.Config)
	}{
		{"missing host", func(c *mysql.Config) { c.Host = "" }},
		{"missing port", func(c *mysql.Config) { c.Port = "" }},
		{"missing database", func(c *mysql.Config) { c.Database = "" }},
		{"missing username", func(c *mysql.Config) { c.Username = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := base
			tt.mutate(&conf)
			if _, err := mysql.DSN(&conf); !errors.Is(err, dbsqlx.ErrMissingConfig) {
				t.Errorf("DSN with %s: err = %v, want ErrMissingConfig", tt.name, err)
			}
		})
	}
}

func TestDSN_PasswordOptional(t *testing.T) {
	dsn, err := mysql.DSN(&mysql.Config{
		Host:     "localhost",
		Port:     "3306",
		Database: "risk",
		Username: "risk",
		// Password intentionally empty — some MySQL setups allow passwordless auth.
	})
	if err != nil {
		t.Fatalf("DSN: %v", err)
	}
	if !strings.HasPrefix(dsn, "risk:@tcp(") {
		t.Errorf("DSN with empty password = %q, want prefix %q", dsn, "risk:@tcp(")
	}
}

// TestNewMySQL_ErrMissingConfig covers the DSN-error path in NewMySQL.
func TestNewMySQL_ErrMissingConfig(t *testing.T) {
	_, err := mysql.NewMySQL(&mysql.Config{}) // all required fields empty
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("NewMySQL empty config: want ErrMissingConfig, got %v", err)
	}
}

// TestInitDefault_Error covers the error-return path in InitDefault.
func TestInitDefault_Error(t *testing.T) {
	err := mysql.InitDefault(&mysql.Config{}) // DSN fails → error propagated
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("InitDefault empty config: want ErrMissingConfig, got %v", err)
	}
}
