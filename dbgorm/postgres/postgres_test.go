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

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dbgorm/postgres"
	gormpostgres "gorm.io/driver/postgres"
)

func TestDialectorFromStructuredFields(t *testing.T) {
	dialector, err := postgres.Dialector(&postgres.Config{
		Host:       "localhost",
		Port:       "5432",
		Database:   "risk",
		Username:   "risk",
		Password:   "secret",
		SearchPath: "public",
	})
	if err != nil {
		t.Fatalf("dialector: %v", err)
	}
	if dialector == nil {
		t.Fatalf("expected dialector")
	}

	pgDialector, ok := dialector.(*gormpostgres.Dialector)
	if !ok {
		t.Fatalf("expected PostgreSQL dialector, got %T", dialector)
	}
	if pgDialector.PreferSimpleProtocol {
		t.Fatal("Dialector must not enable simple protocol")
	}
	if strings.Contains(pgDialector.DSN, "default_query_exec_mode=") {
		t.Fatalf("Dialector DSN must preserve pgx default, got %q", pgDialector.DSN)
	}
}

func TestDialectorQueryExecMode(t *testing.T) {
	base := postgres.Config{
		Host:     "localhost",
		Port:     "5432",
		Database: "risk",
		Username: "risk",
		Password: "secret",
	}
	tests := []struct {
		name    string
		mode    string
		want    string
		wantErr bool
	}{
		{name: "pgx default"},
		{name: "cache statement", mode: "cache_statement", want: "cache_statement"},
		{name: "cache describe", mode: "cache_describe", want: "cache_describe"},
		{name: "describe exec", mode: "describe_exec", want: "describe_exec"},
		{name: "simple protocol", mode: "simple_protocol", want: "simple_protocol"},
		{name: "invalid", mode: "invalid", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := base
			conf.QueryExecMode = tt.mode
			dialector, err := postgres.Dialector(&conf)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Dialector should reject the query execution mode")
				}
				return
			}
			if err != nil {
				t.Fatalf("Dialector: %v", err)
			}
			pgDialector := dialector.(*gormpostgres.Dialector)
			if tt.want == "" {
				if strings.Contains(pgDialector.DSN, "default_query_exec_mode=") {
					t.Fatalf("DSN must preserve pgx default, got %q", pgDialector.DSN)
				}
				return
			}
			if !strings.Contains(pgDialector.DSN, "default_query_exec_mode="+tt.want) {
				t.Fatalf("DSN = %q, want query mode %q", pgDialector.DSN, tt.want)
			}
		})
	}
}

func TestDialectorRequiresStructuredFields(t *testing.T) {
	if _, err := postgres.Dialector(&postgres.Config{}); err != dbgorm.ErrMissingConfig {
		t.Fatalf("expected ErrMissingConfig, got %v", err)
	}
}

func TestDialectorDoesNotConnect(t *testing.T) {
	dialector, err := postgres.Dialector(&postgres.Config{
		Host:     "localhost",
		Port:     "5432",
		Database: "risk",
		Username: "risk",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("dialector: %v", err)
	}
	if dialector == nil {
		t.Fatalf("expected dialector")
	}
}

// TestNewPostgres_ErrMissingConfig covers the Dialector-error path in NewPostgres.
func TestNewPostgres_ErrMissingConfig(t *testing.T) {
	_, err := postgres.NewPostgres(&postgres.Config{}) // all required fields empty
	if !errors.Is(err, dbgorm.ErrMissingConfig) {
		t.Errorf("NewPostgres empty config: want ErrMissingConfig, got %v", err)
	}
}

// TestInitDefault_Error covers the error-return path in InitDefault.
func TestInitDefault_Error(t *testing.T) {
	err := postgres.InitDefault(&postgres.Config{}) // Dialector fails → error propagated
	if !errors.Is(err, dbgorm.ErrMissingConfig) {
		t.Errorf("InitDefault empty config: want ErrMissingConfig, got %v", err)
	}
}
