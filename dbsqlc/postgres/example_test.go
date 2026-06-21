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
	"context"
	"fmt"

	"github.com/phcp-tech/common-library-golang/dbsqlc/postgres"
	"github.com/phcp-tech/common-library-golang/health"
)

// ExampleNewPostgres shows the typical usage of NewPostgres.
// NewPostgres performs an eager connectivity check (show search_path) so callers
// discover unreachable databases immediately at startup rather than on the first
// real query. Zero-value pool fields fall back to the dbsqlc package defaults
// (MaxOpenConns=100, MaxIdleConns=25, ConnMaxLifetime=60min, ConnMaxIdletime=10min).
func ExampleNewPostgres() {
	// With an unreachable host the connectivity check fails and an error is returned.
	_, err := postgres.NewPostgres(&postgres.Config{
		Host:     "127.0.0.1",
		Port:     "19999",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	fmt.Println(err != nil) // true — connectivity check failed
	// Output:
	// true
}

// ExampleNewPostgres_customPool shows how to override the default connection pool settings.
func ExampleNewPostgres_customPool() {
	_, err := postgres.NewPostgres(&postgres.Config{
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
	fmt.Println(err != nil) // true — connectivity check failed
	// Output:
	// true
}

// ExampleInitDefault shows the singleton pattern for the default PostgreSQL pool.
// Call InitDefault once at application startup; subsequent calls are silently
// ignored (sync.Once).
//
// In this example the singleton was already consumed by a prior test in this
// binary, so InitDefault is a no-op and returns nil. Default() returns nil
// because the first initialisation attempt failed (unreachable host).
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
	fmt.Println(err != nil)                // false — sync.Once no-op, returns nil
	fmt.Println(postgres.Default() != nil) // false — first init failed, instance nil
	// Output:
	// false
	// false
}

// ExampleHealthChecker shows how to wire HealthChecker into a [health.Check] call
// for a /health HTTP endpoint.
// When the default pool has not been initialised, the Checker reports StatusUnhealthy.
func ExampleHealthChecker() {
	results := health.Check(context.Background(), postgres.HealthChecker())
	fmt.Println(results[0].Name)
	fmt.Println(results[0].Status == health.StatusUnhealthy) // true — Default() is nil
	// Output:
	// database
	// true
}
