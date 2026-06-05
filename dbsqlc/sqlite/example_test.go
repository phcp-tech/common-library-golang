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

	"github.com/phcp-tech/common-library-golang/dbsqlc/sqlite"
)

// ExampleNewSQLite_memory shows how to open an ephemeral in-memory SQLite database.
// In-memory databases are lost when the *sql.DB is closed; they are ideal for tests
// and short-lived operations that require no persistence.
func ExampleNewSQLite_memory() {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer db.Close()
	fmt.Println("opened")
	// Output:
	// opened
}

// ExampleNewSQLite_file shows URI query parameters for enabling WAL mode and
// foreign key enforcement on a file-based SQLite database.
// WAL mode allows concurrent readers while a write is in progress.
func ExampleNewSQLite_file() {
	db, err := sqlite.NewSQLite(&sqlite.Config{
		Path: "file::memory:?_journal_mode=WAL&_foreign_keys=on",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer db.Close()
	fmt.Println("opened")
	// Output:
	// opened
}

// ExampleInitDefault shows the singleton pattern: call InitDefault once at
// application startup. Subsequent calls are silently ignored (sync.Once).
func ExampleInitDefault() {
	err := sqlite.InitDefault(&sqlite.Config{Path: ":memory:"})
	fmt.Println(err)
	// Output:
	// <nil>
}

// ExampleDefault shows how to retrieve the singleton *sql.DB created by InitDefault.
// Returns nil if InitDefault has not been called yet.
func ExampleDefault() {
	db := sqlite.Default()
	_ = db // pass to sqlc-generated Queries
}
