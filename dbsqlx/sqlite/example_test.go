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

// package sqlite_test demonstrates the public API from a caller's perspective.
package sqlite_test

import (
	"fmt"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dbsqlx/sqlite"
)

// ExampleNewSQLite shows how to open an in-memory SQLite database.
// In-memory databases are useful for tests and ephemeral data — all data is
// lost when the connection is closed.
func ExampleNewSQLite() {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	fmt.Println(err == nil)
	fmt.Println(db != nil)
	// Output:
	// true
	// true
}

// ExampleInitDefault shows the default-instance pattern: call InitDefault once
// at application startup, then retrieve the shared *sqlx.DB via dbsqlx.Default().
func ExampleInitDefault() {
	err := sqlite.InitDefault(&sqlite.Config{Path: ":memory:"})
	fmt.Println(err == nil)
	fmt.Println(dbsqlx.Default() != nil)
	// Output:
	// true
	// true
}
