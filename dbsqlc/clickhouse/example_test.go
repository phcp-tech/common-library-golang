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

// package clickhouse_test demonstrates the public API from a caller's perspective.
package clickhouse_test

import (
	"context"
	"fmt"

	"github.com/phcp-tech/common-library-golang/dbsqlc/clickhouse"
	"github.com/phcp-tech/common-library-golang/health"
)

// ExampleNewClickHouse shows how to open a ClickHouse connection with the
// default pool settings.
// Zero-value pool fields fall back to the dbsqlc package defaults
// (MaxOpenConns=100, MaxIdleConns=25, ConnMaxLifetime=60min).
// Note: unlike MySQL/PostgreSQL, ClickHouse establishes a real TCP connection
// at Open time, so a live server is required.
func ExampleNewClickHouse() {
	conn, err := clickhouse.NewClickHouse(&clickhouse.Config{
		Host:     "127.0.0.1",
		Port:     "9000",
		Database: "default",
		Username: "default",
		Password: "",
	})
	if err != nil {
		// handle error
		return
	}
	defer conn.Close()
}

// ExampleNewClickHouse_customPool shows how to override the default connection
// pool settings.
func ExampleNewClickHouse_customPool() {
	conn, err := clickhouse.NewClickHouse(&clickhouse.Config{
		Host:            "127.0.0.1",
		Port:            "9000",
		Database:        "default",
		Username:        "default",
		Password:        "",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30, // minutes
	})
	if err != nil {
		// handle error
		return
	}
	defer conn.Close()
}

// ExampleInitDefault shows the singleton pattern: call InitDefault once at
// application startup. Subsequent calls are silently ignored (sync.Once).
// After a successful call, Default returns the initialised driver.Conn which
// can be passed directly to sqlc-generated Queries.
func ExampleInitDefault() {
	if err := clickhouse.InitDefault(&clickhouse.Config{
		Host:     "127.0.0.1",
		Port:     "9000",
		Database: "default",
		Username: "default",
		Password: "",
	}); err != nil {
		// handle error
		return
	}
	_ = clickhouse.Default() // driver.Conn, pass to sqlc-generated Queries
}

// ExampleHealthChecker shows how to wire HealthChecker into a [health.Check] call
// for a /health HTTP endpoint.
// Default() is nil in this example so the Checker reports StatusUnhealthy.
func ExampleHealthChecker() {
	results := health.Check(context.Background(), clickhouse.HealthChecker())
	fmt.Println(results[0].Name)
	fmt.Println(results[0].Status == health.StatusUnhealthy) // true — Default() is nil
	// Output:
	// clickhouse
	// true
}
