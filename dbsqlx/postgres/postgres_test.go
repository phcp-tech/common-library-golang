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

package postgres_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dbsqlx/postgres"
)

func TestDSN_FromStructuredFields(t *testing.T) {
	dsn, err := postgres.DSN(&postgres.Config{
		Host:       "localhost",
		Port:       "5432",
		Database:   "risk",
		Username:   "risk",
		Password:   "secret",
		SearchPath: "public",
	})
	if err != nil {
		t.Fatalf("DSN: %v", err)
	}
	if dsn == "" {
		t.Fatal("expected non-empty dsn")
	}
	for _, want := range []string{"host=localhost", "port=5432", "user=risk", "dbname=risk", "search_path=public"} {
		if !strings.Contains(dsn, want) {
			t.Errorf("dsn = %q, want it to contain %q", dsn, want)
		}
	}
}

func TestDSN_OmitsSearchPathWhenEmpty(t *testing.T) {
	dsn, err := postgres.DSN(&postgres.Config{
		Host:     "localhost",
		Port:     "5432",
		Database: "risk",
		Username: "risk",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("DSN: %v", err)
	}
	if strings.Contains(dsn, "search_path") {
		t.Errorf("dsn = %q, want no search_path segment", dsn)
	}
}

func TestDSN_RequiresStructuredFields(t *testing.T) {
	if _, err := postgres.DSN(&postgres.Config{}); !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Fatalf("DSN empty config: err = %v, want ErrMissingConfig", err)
	}
}

// TestNewPostgres_ErrMissingConfig covers the DSN-error path in NewPostgres.
func TestNewPostgres_ErrMissingConfig(t *testing.T) {
	_, err := postgres.NewPostgres(&postgres.Config{}) // all required fields empty
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("NewPostgres empty config: want ErrMissingConfig, got %v", err)
	}
}

// TestInitDefault_Error covers the error-return path in InitDefault.
func TestInitDefault_Error(t *testing.T) {
	err := postgres.InitDefault(&postgres.Config{}) // DSN fails → error propagated
	if !errors.Is(err, dbsqlx.ErrMissingConfig) {
		t.Errorf("InitDefault empty config: want ErrMissingConfig, got %v", err)
	}
}
