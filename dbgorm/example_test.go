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

package dbgorm_test

import (
	"context"
	"fmt"

	libgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dbgorm/sqlite"
	"github.com/phcp-tech/common-library-golang/health"
	"gorm.io/gorm"
)

// exProduct is a minimal GORM model used by the example functions.
type exProduct struct {
	ID       uint `gorm:"primaryKey"`
	Name     string
	Category string
}

// exDB opens a private in-memory SQLite database and migrates exProduct.
// Each call returns an independent database — no shared state between examples.
func exDB() *gorm.DB {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		panic("exDB: " + err.Error())
	}
	if err := db.AutoMigrate(&exProduct{}); err != nil {
		panic("exDB migrate: " + err.Error())
	}
	return db
}

// -----------------------------------------------------------------------
// Open / Close
// -----------------------------------------------------------------------

// ExampleOpen shows the low-level Open API using a GORM dialector directly.
// Prefer the adapter packages (dbgorm/mysql, dbgorm/postgres, dbgorm/sqlite)
// for application code.
func ExampleOpen() {
	dialector, _ := sqlite.Dialector(&sqlite.Config{Path: ":memory:"})
	db, err := libgorm.Open(dialector, &libgorm.GormConfig{
		MaxOpenConns: 1,
		MaxIdleConns: 1,
	})
	fmt.Println(err == nil)
	fmt.Println(db != nil)
	// Output:
	// true
	// true
}

// ExampleClose shows how to close a GORM database obtained from an adapter.
// Close is a no-op when db is nil.
func ExampleClose() {
	db := exDB()
	err := libgorm.Close(db)
	fmt.Println(err == nil)
	// Output:
	// true
}

// ExampleIsNotFound shows how to distinguish a not-found error from other
// database errors. Use it instead of comparing directly against gorm.ErrRecordNotFound.
func ExampleIsNotFound() {
	ctx := context.Background()
	db := exDB()

	_, err := libgorm.FirstByID[exProduct](ctx, db, 9999)
	fmt.Println(libgorm.IsNotFound(err)) // true — no row with id=9999
	// Output:
	// true
}

// -----------------------------------------------------------------------
// Default / SetDefault
// -----------------------------------------------------------------------

// ExampleSetDefault shows how to store a *gorm.DB as the process-wide default.
// Call SetDefault once at application startup after opening the database.
func ExampleSetDefault() {
	prev := libgorm.Default()
	defer libgorm.SetDefault(prev) // restore after example

	libgorm.SetDefault(exDB())
	fmt.Println(libgorm.Default() != nil)
	// Output:
	// true
}

// ExampleDefault shows how to retrieve the process-wide default *gorm.DB.
func ExampleDefault() {
	prev := libgorm.Default()
	libgorm.SetDefault(nil)
	defer libgorm.SetDefault(prev) // restore after example

	fmt.Println(libgorm.Default() == nil) // nil — not yet initialised
	// Output:
	// true
}

// -----------------------------------------------------------------------
// HealthChecker
// -----------------------------------------------------------------------

// ExampleHealthChecker shows how to wire HealthChecker into a health.Check
// call for a /health HTTP endpoint.
func ExampleHealthChecker() {
	prev := libgorm.Default()
	libgorm.SetDefault(nil)
	defer libgorm.SetDefault(prev)

	results := health.Check(context.Background(), libgorm.HealthChecker())
	fmt.Println(results[0].Name)
	fmt.Println(results[0].Status == health.StatusUnhealthy) // true — Default() is nil
	// Output:
	// database
	// true
}

// -----------------------------------------------------------------------
// Query helpers
// -----------------------------------------------------------------------

// ExampleFirstByID shows how to fetch a record by primary key.
// Returns gorm.ErrRecordNotFound when no row matches.
func ExampleFirstByID() {
	ctx := context.Background()
	db := exDB()
	p := &exProduct{Name: "widget", Category: "tools"}
	db.Create(p)

	found, err := libgorm.FirstByID[exProduct](ctx, db, p.ID)
	fmt.Println(err == nil)
	fmt.Println(found.Name)
	// Output:
	// true
	// widget
}

// ExampleFirstWhere shows how to fetch the first record matching a condition.
func ExampleFirstWhere() {
	ctx := context.Background()
	db := exDB()
	db.Create(&exProduct{Name: "hammer", Category: "tools"})
	db.Create(&exProduct{Name: "wrench", Category: "tools"})

	found, err := libgorm.FirstWhere[exProduct](ctx, db, "name = ?", "wrench")
	fmt.Println(err == nil)
	fmt.Println(found.Name)
	// Output:
	// true
	// wrench
}

// ExampleDeleteByID shows how to delete a record by primary key.
// Returns gorm.ErrRecordNotFound when no row is deleted.
func ExampleDeleteByID() {
	ctx := context.Background()
	db := exDB()
	p := &exProduct{Name: "bolt"}
	db.Create(p)

	err := libgorm.DeleteByID[exProduct](ctx, db, p.ID)
	fmt.Println(err == nil)

	// Deleting again returns not-found.
	fmt.Println(libgorm.IsNotFound(libgorm.DeleteByID[exProduct](ctx, db, p.ID)))
	// Output:
	// true
	// true
}

// ExampleDeleteWhere shows how to delete all records matching a condition.
// Zero affected rows is treated as success.
func ExampleDeleteWhere() {
	ctx := context.Background()
	db := exDB()
	db.Create(&exProduct{Name: "nail", Category: "fasteners"})
	db.Create(&exProduct{Name: "screw", Category: "fasteners"})

	err := libgorm.DeleteWhere(ctx, db, &exProduct{}, "category = ?", "fasteners")
	fmt.Println(err == nil)
	// Output:
	// true
}

// -----------------------------------------------------------------------
// Scopes — Paginate / OrderBy
// -----------------------------------------------------------------------

// ExamplePaginate shows how to apply limit/offset pagination as a GORM scope.
func ExamplePaginate() {
	ctx := context.Background()
	db := exDB()
	for i := range 5 {
		db.Create(&exProduct{Name: fmt.Sprintf("item-%d", i)})
	}

	var page []exProduct
	db.WithContext(ctx).Scopes(libgorm.Paginate(1, 3)).Find(&page)
	fmt.Println(len(page)) // page 1, limit 3
	// Output:
	// 3
}

// ExampleOrderBy shows how to apply an allow-listed column sort as a GORM scope.
func ExampleOrderBy() {
	ctx := context.Background()
	db := exDB()
	db.Create(&exProduct{Name: "zebra"})
	db.Create(&exProduct{Name: "apple"})
	db.Create(&exProduct{Name: "mango"})

	allowed := map[string]string{"name": "name"}
	var results []exProduct
	db.WithContext(ctx).Scopes(libgorm.OrderBy(allowed, "name", "ASC")).Find(&results)
	fmt.Println(results[0].Name)
	fmt.Println(results[2].Name)
	// Output:
	// apple
	// zebra
}

// -----------------------------------------------------------------------
// Raw SQL
// -----------------------------------------------------------------------

// ExampleExecRaw shows how to execute a raw SQL statement with positional args.
func ExampleExecRaw() {
	ctx := context.Background()
	db := exDB()
	p := &exProduct{Name: "gear", Category: "parts"}
	db.Create(p)

	rows, err := libgorm.ExecRaw(ctx, db,
		"UPDATE ex_products SET category = ? WHERE id = ?", "machinery", p.ID)
	fmt.Println(err == nil)
	fmt.Println(rows) // 1 row updated
	// Output:
	// true
	// 1
}

// ExampleScanRaw shows how to scan a scalar result (e.g. COUNT) from a raw query.
func ExampleScanRaw() {
	ctx := context.Background()
	db := exDB()
	db.Create(&exProduct{Name: "pin", Category: "fasteners"})
	db.Create(&exProduct{Name: "clip", Category: "fasteners"})

	var total int64
	err := libgorm.ScanRaw(ctx, db, &total,
		"SELECT COUNT(*) FROM ex_products WHERE category = ?", "fasteners")
	fmt.Println(err == nil)
	fmt.Println(total)
	// Output:
	// true
	// 2
}

// ExampleScanRaw_singleRow shows how to scan the first matching row into a struct.
// Only the first row is populated; additional rows are discarded.
func ExampleScanRaw_singleRow() {
	ctx := context.Background()
	db := exDB()
	db.Create(&exProduct{Name: "hammer", Category: "tools"})

	var p exProduct
	err := libgorm.ScanRaw(ctx, db, &p,
		"SELECT * FROM ex_products WHERE name = ?", "hammer")
	fmt.Println(err == nil)
	fmt.Println(p.Name)
	// Output:
	// true
	// hammer
}

// ExampleScanRaw_multipleRows shows how to scan all matching rows into a slice.
// Pass a pointer to a slice as dest to collect every returned row.
func ExampleScanRaw_multipleRows() {
	ctx := context.Background()
	db := exDB()
	db.Create(&exProduct{Name: "nail", Category: "fasteners"})
	db.Create(&exProduct{Name: "bolt", Category: "fasteners"})
	db.Create(&exProduct{Name: "screw", Category: "fasteners"})

	var products []exProduct
	err := libgorm.ScanRaw(ctx, db, &products,
		"SELECT * FROM ex_products WHERE category = ? ORDER BY name", "fasteners")
	fmt.Println(err == nil)
	fmt.Println(len(products))
	fmt.Println(products[0].Name)
	// Output:
	// true
	// 3
	// bolt
}
