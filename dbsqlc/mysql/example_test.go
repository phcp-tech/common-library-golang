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

// package mysql_test demonstrates the public API from a caller's perspective.
package mysql_test

import (
	"context"
	"fmt"

	"github.com/phcp-tech/common-library-golang/dbsqlc/mysql"
	"github.com/phcp-tech/common-library-golang/health"
)

// ExampleNewMySQL shows how to open a MySQL connection with the default pool settings.
// sql.Open is lazy — no connection is established until the database is first used.
// For cases that require multiple connections use NewMySQL directly instead of the singleton.
func ExampleNewMySQL() {
	db, err := mysql.NewMySQL(&mysql.Config{
		Host:     "127.0.0.1",
		Port:     "3306",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer db.Close()
	fmt.Println(db != nil)
}

// ExampleNewMySQL_customPool shows how to override the default connection pool settings.
// Zero-value pool fields fall back to the dbsqlc package defaults
// (MaxOpenConns=100, MaxIdleConns=25, ConnMaxLifetime=60min, ConnMaxIdletime=10min).
func ExampleNewMySQL_customPool() {
	db, err := mysql.NewMySQL(&mysql.Config{
		Host:            "127.0.0.1",
		Port:            "3306",
		Database:        "mydb",
		Username:        "user",
		Password:        "pass",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30, // minutes
		ConnMaxIdletime: 5,  // minutes
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer db.Close()
	fmt.Println(db != nil)
}

// ExampleInitDefault shows the singleton pattern: call InitDefault once at
// application startup. Subsequent calls are silently ignored (sync.Once).
// After a successful call, Default returns the initialised *sql.DB which can
// be passed directly to sqlc-generated Queries.
func ExampleInitDefault() {
	err := mysql.InitDefault(&mysql.Config{
		Host:     "127.0.0.1",
		Port:     "3306",
		Database: "mydb",
		Username: "user",
		Password: "pass",
	})
	fmt.Println(err == nil)
	fmt.Println(mysql.Default() != nil)
}

// ExampleHealthChecker shows how to wire HealthChecker into a [health.Check] call
// for a /health HTTP endpoint.
// In test environments with no reachable MySQL server, the Checker reports StatusUnhealthy
// regardless of whether the singleton has been initialised.
func ExampleHealthChecker() {
	results := health.Check(context.Background(), mysql.HealthChecker())
	fmt.Println(results[0].Name)
	fmt.Println(results[0].Status == health.StatusUnhealthy) // true — no reachable MySQL
	// Output:
	// database
	// true
}
