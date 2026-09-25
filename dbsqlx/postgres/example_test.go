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

// package postgres_test demonstrates the public API from a caller's perspective.
package postgres_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx/postgres"
	"github.com/phcp-tech/common-library-golang/dto"
)

// ExampleNewPostgres shows how to open a PostgreSQL-backed *sqlx.DB.
// Open verifies connectivity with an eager ping and runs SHOW search_path —
// a non-nil error is returned immediately when the server is unreachable.
func ExampleNewPostgres() {
	_, err := postgres.NewPostgres(&postgres.Config{
		Host:     "127.0.0.1",
		Port:     "1",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	fmt.Println(err != nil) // true — server unreachable, Open pings eagerly
	// Output:
	// true
}

func TestDSNQueryExecMode(t *testing.T) {
	base := postgres.Config{
		Host:     "localhost",
		Port:     "5432",
		Database: "app",
		Username: "user",
		Password: "pass",
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
			dsn, err := postgres.DSN(&conf)
			if tt.wantErr {
				if err == nil {
					t.Fatal("DSN should reject the query execution mode")
				}
				return
			}
			if err != nil {
				t.Fatalf("DSN: %v", err)
			}
			if tt.want == "" {
				if strings.Contains(dsn, "default_query_exec_mode=") {
					t.Fatalf("DSN must preserve pgx default, got %q", dsn)
				}
				return
			}
			if !strings.Contains(dsn, "default_query_exec_mode="+tt.want) {
				t.Fatalf("DSN = %q, want query mode %q", dsn, tt.want)
			}
		})
	}
}

// ExampleDSN_queryExecMode shows how to select a non-default pgx execution mode.
// See the package documentation before using a mode other than exec.
func ExampleDSN_queryExecMode() {
	dsn, err := postgres.DSN(&postgres.Config{
		Host:          "localhost",
		Port:          "5432",
		Database:      "mydb",
		Username:      "user",
		Password:      "pass",
		QueryExecMode: postgres.QueryExecModeCacheStatement,
	})
	fmt.Println(err)
	fmt.Println(strings.Contains(dsn, "default_query_exec_mode=cache_statement"))
	// Output:
	// <nil>
	// true
}

// ExampleInitDefault shows the default-instance pattern.
// Call InitDefault once at application startup; dbsqlx.Default() returns
// the shared *sqlx.DB for the process.
func ExampleInitDefault() {
	err := postgres.InitDefault(&postgres.Config{
		Host:     "127.0.0.1",
		Port:     "1",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	fmt.Println(err != nil) // true — server unreachable
}

// ExampleDSN shows how to build a libpq-style connection string, including an
// optional search_path.
func ExampleDSN() {
	dsn, err := postgres.DSN(&postgres.Config{
		Host:       "localhost",
		Port:       "5432",
		Database:   "mydb",
		Username:   "user",
		Password:   "pass",
		SearchPath: "myschema",
	})
	fmt.Println(err == nil)
	fmt.Println(dsn)
	// Output:
	// true
	// host=localhost port=5432 user=user password=pass dbname=mydb sslmode=disable TimeZone=UTC search_path=myschema
}

// ExampleZhSortSql shows how to build an ORDER BY clause that
// approximates pinyin ordering for a Chinese-text column, then append it
// directly to a raw SELECT statement alongside dbsqlx.PageSql for pagination.
// This is PostgreSQL-specific — see ZhSortSql's doc comment for why.
func ExampleZhSortSql() {
	para := dto.PageParameter{Sort: "name", Direction: "ASC"}
	query := "SELECT * FROM products" + postgres.ZhSortSql(&para)
	fmt.Println(query)
	// Output:
	// SELECT * FROM products ORDER BY convert_to(name,'GBK') ASC
}
