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
	"context"
	"fmt"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dto"
	"github.com/phcp-tech/common-library-golang/health"

	"github.com/vinovest/sqlx"
	_ "modernc.org/sqlite"
)

// exProduct is a minimal model used by the example functions.
type exProduct struct {
	ID       int64  `db:"id"`
	Name     string `db:"name"`
	Category string `db:"category"`
}

// exDB opens a private in-memory SQLite database and creates the ex_products
// table. Each call returns an independent database — no shared state between
// examples. Uses the plain sqlite driver directly (not the dbsqlx/sqlite
// adapter) to keep these examples focused on the driver-agnostic root API.
func exDB() *sqlx.DB {
	db, err := dbsqlx.Open("sqlite", ":memory:", nil)
	if err != nil {
		panic("exDB: " + err.Error())
	}
	if _, err := db.Exec(`CREATE TABLE ex_products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		category TEXT NOT NULL DEFAULT ''
	)`); err != nil {
		panic("exDB create table: " + err.Error())
	}
	return db
}

// -----------------------------------------------------------------------
// Open / Close
// -----------------------------------------------------------------------

// ExampleOpen shows how to connect to a database with pool tuning.
// Open verifies connectivity with an eager ping before returning.
func ExampleOpen() {
	db, err := dbsqlx.Open("sqlite", ":memory:", &dbsqlx.PoolConfig{
		MaxOpenConns: 4,
		MaxIdleConns: 2,
	})
	fmt.Println(err == nil)
	fmt.Println(db != nil)
	// Output:
	// true
	// true
}

// ExampleClose shows how to close a *sqlx.DB obtained from Open.
// Close is a no-op when db is nil.
func ExampleClose() {
	db := exDB()
	err := dbsqlx.Close(db)
	fmt.Println(err == nil)
	// Output:
	// true
}

// -----------------------------------------------------------------------
// Default / SetDefault
// -----------------------------------------------------------------------

// ExampleSetDefault shows how to store a *sqlx.DB as the process-wide default.
// Call SetDefault once at application startup after opening the database.
func ExampleSetDefault() {
	prev := dbsqlx.Default()
	defer dbsqlx.SetDefault(prev) // restore after example

	dbsqlx.SetDefault(exDB())
	fmt.Println(dbsqlx.Default() != nil)
	// Output:
	// true
}

// ExampleDefault shows how to retrieve the process-wide default *sqlx.DB.
func ExampleDefault() {
	prev := dbsqlx.Default()
	dbsqlx.SetDefault(nil)
	defer dbsqlx.SetDefault(prev) // restore after example

	fmt.Println(dbsqlx.Default() == nil) // nil — not yet initialised
	// Output:
	// true
}

// -----------------------------------------------------------------------
// HealthChecker
// -----------------------------------------------------------------------

// ExampleHealthChecker shows how to wire HealthChecker into a health.Check
// call for a /health HTTP endpoint.
func ExampleHealthChecker() {
	prev := dbsqlx.Default()
	dbsqlx.SetDefault(nil)
	defer dbsqlx.SetDefault(prev)

	results := health.Check(context.Background(), dbsqlx.HealthChecker())
	fmt.Println(results[0].Name)
	fmt.Println(results[0].Status == health.StatusUnhealthy) // true — Default() is nil
	// Output:
	// database
	// true
}

// -----------------------------------------------------------------------
// Exec
// -----------------------------------------------------------------------

// ExampleExec shows how to run a statement and get the number of affected rows.
func ExampleExec() {
	ctx := context.Background()
	db := exDB()
	db.MustExec(`INSERT INTO ex_products (name, category) VALUES (?, ?)`, "nail", "fasteners")
	db.MustExec(`INSERT INTO ex_products (name, category) VALUES (?, ?)`, "screw", "fasteners")

	rows, err := dbsqlx.Exec(ctx, db, `UPDATE ex_products SET category = ? WHERE category = ?`, "hardware", "fasteners")
	fmt.Println(err == nil)
	fmt.Println(rows)
	// Output:
	// true
	// 2
}

// -----------------------------------------------------------------------
// NamedExec
// -----------------------------------------------------------------------

// ExampleNamedExec shows how to run a statement with named parameters bound
// from a map. Named parameters improve readability for statements with many
// columns.
func ExampleNamedExec() {
	ctx := context.Background()
	db := exDB()

	rows, err := dbsqlx.NamedExec(ctx, db,
		`INSERT INTO ex_products (name, category) VALUES (:name, :category)`,
		map[string]any{"name": "chisel", "category": "tools"})
	fmt.Println(err == nil)
	fmt.Println(rows)
	// Output:
	// true
	// 1
}

// ExampleNamedExec_struct shows how to bind named parameters directly from a
// struct's db-tagged fields instead of a map.
func ExampleNamedExec_struct() {
	ctx := context.Background()
	db := exDB()

	arg := exProduct{Name: "saw", Category: "tools"}
	rows, err := dbsqlx.NamedExec(ctx, db,
		`INSERT INTO ex_products (name, category) VALUES (:name, :category)`, arg)
	fmt.Println(err == nil)
	fmt.Println(rows)
	// Output:
	// true
	// 1
}

// -----------------------------------------------------------------------
// Transact
// -----------------------------------------------------------------------

// ExampleTransact shows how to run several statements atomically.
// The callback receives tx (a sqlx.Queryable) instead of db — pass tx, not
// db, to every helper called inside the callback.
func ExampleTransact() {
	ctx := context.Background()
	db := exDB()

	err := dbsqlx.Transact(ctx, db, func(ctx context.Context, tx sqlx.Queryable) error {
		if _, err := dbsqlx.Exec(ctx, tx, `INSERT INTO ex_products (name) VALUES (?)`, "drill"); err != nil {
			return err
		}
		_, err := dbsqlx.Exec(ctx, tx, `INSERT INTO ex_products (name) VALUES (?)`, "sander")
		return err
	})
	fmt.Println(err == nil)

	var count int64
	_ = db.GetContext(ctx, &count, `SELECT COUNT(*) FROM ex_products`)
	fmt.Println(count)
	// Output:
	// true
	// 2
}

// ExampleTransact_rollback shows that a non-nil error returned from the
// callback rolls back every statement executed inside the transaction.
func ExampleTransact_rollback() {
	ctx := context.Background()
	db := exDB()

	err := dbsqlx.Transact(ctx, db, func(ctx context.Context, tx sqlx.Queryable) error {
		if _, err := dbsqlx.Exec(ctx, tx, `INSERT INTO ex_products (name) VALUES (?)`, "temp"); err != nil {
			return err
		}
		return fmt.Errorf("something went wrong")
	})
	fmt.Println(err)

	var count int64
	_ = db.GetContext(ctx, &count, `SELECT COUNT(*) FROM ex_products`)
	fmt.Println(count) // 0 — the insert was rolled back
	// Output:
	// something went wrong
	// 0
}

// -----------------------------------------------------------------------
// SortSql / PageSql
// -----------------------------------------------------------------------

// ExampleSortSql shows how to build a safe ORDER BY clause from user-supplied
// sort parameters and append it directly to a raw SELECT statement.
func ExampleSortSql() {
	ctx := context.Background()
	db := exDB()
	db.MustExec(`INSERT INTO ex_products (name) VALUES (?)`, "zebra")
	db.MustExec(`INSERT INTO ex_products (name) VALUES (?)`, "apple")

	para := dto.PageParameter{Sort: "name", Direction: "ASC"}
	query := "SELECT * FROM ex_products" + dbsqlx.SortSql(&para)

	var products []exProduct
	err := db.SelectContext(ctx, &products, query)
	fmt.Println(err == nil)
	fmt.Println(products[0].Name)
	// Output:
	// true
	// apple
}

// ExamplePageSql shows how to build a LIMIT/OFFSET clause and append it to a
// raw SELECT statement for pagination.
func ExamplePageSql() {
	ctx := context.Background()
	db := exDB()
	for i := range 5 {
		db.MustExec(`INSERT INTO ex_products (name) VALUES (?)`, fmt.Sprintf("item-%d", i))
	}

	para := dto.PageParameter{Page: 1, Limit: 3}
	query := "SELECT * FROM ex_products" + dbsqlx.PageSql(&para)

	var products []exProduct
	err := db.SelectContext(ctx, &products, query)
	fmt.Println(err == nil)
	fmt.Println(len(products)) // page 1, limit 3
	// Output:
	// true
	// 3
}
