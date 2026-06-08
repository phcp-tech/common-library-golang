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

	"github.com/phcp-tech/common-library-golang/dbsqlc/postgres"
)

// ExampleNewPostgres shows that pool creation is lazy: pgxpool.NewWithConfig does
// not establish any connections at creation time, so no live PostgreSQL server is
// required. Connections are acquired on first use.
// Zero-value pool fields fall back to the dbsqlc package defaults
// (MaxOpenConns=100, MaxIdleConns=25, ConnMaxLifetime=60min, ConnMaxIdletime=10min).
func ExampleNewPostgres() {
	pool, err := postgres.NewPostgres(&postgres.Config{
		Host:     "127.0.0.1",
		Port:     "19999",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer pool.Close()
	fmt.Println("pool created")
	// Output:
	// pool created
}

// ExampleNewPostgres_customPool shows how to override the default connection pool settings.
func ExampleNewPostgres_customPool() {
	pool, err := postgres.NewPostgres(&postgres.Config{
		Host:            "127.0.0.1",
		Port:            "19999",
		Database:        "mydb",
		Username:        "user",
		Password:        "pass",
		SearchPath:      "myschema",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30, // minutes
		ConnMaxIdletime: 5,  // minutes
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer pool.Close()
	fmt.Println("pool created")
	// Output:
	// pool created
}

// ExampleInitDefault shows the singleton pattern for the default PostgreSQL pool.
// Call InitDefault once at application startup; subsequent calls are silently
// ignored (sync.Once). After a successful call, Default returns the initialised
// *pgxpool.Pool which can be passed directly to sqlc-generated Queries.
func ExampleInitDefault() {
	err := postgres.InitDefault(&postgres.Config{
		Host:            "127.0.0.1",
		Port:            "19999",
		Database:        "mydb",
		Username:        "user",
		Password:        "pass",
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 60,
		ConnMaxIdletime: 10,
	})
	fmt.Println(err)
	fmt.Println(postgres.Default() != nil)
	// Output:
	// <nil>
	// true
}
