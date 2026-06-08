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

// package vfs_test demonstrates the public API from a caller's perspective.
package vfs_test

import (
	"embed"
	"fmt"

	sqlitevfs "github.com/phcp-tech/common-library-golang/dbsqlc/sqlite/vfs"
)

// ExampleNew shows how to open an embedded SQLite database via VFS.
// The database file must be embedded at "config/sqlite.db" within the FS.
//
// In production code, replace the empty embed.FS declaration with:
//
//	//go:embed config/sqlite.db
//	var sqliteFS embed.FS
func ExampleNew() {
	var sqliteFS embed.FS // placeholder — embed with //go:embed config/sqlite.db

	db, err := sqlitevfs.New(&sqliteFS)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer db.Close()

	// pass db to sqlc-generated Queries
	_ = db
}

// ExampleInitDefault shows the singleton pattern: call InitDefault once at
// application startup. Subsequent calls are silently ignored (sync.Once).
// After a successful call, Default returns the initialised *sql.DB.
//
// In production code, replace the empty embed.FS declaration with:
//
//	//go:embed config/sqlite.db
//	var sqliteFS embed.FS
func ExampleInitDefault() {
	var sqliteFS embed.FS // placeholder — embed with //go:embed config/sqlite.db

	if err := sqlitevfs.InitDefault(&sqliteFS); err != nil {
		fmt.Println("error:", err)
		return
	}

	db := sqlitevfs.Default() // *sql.DB, pass to sqlc-generated Queries
	_ = db
}

// ExampleDefault shows how to retrieve the singleton *sql.DB created by InitDefault.
// Returns nil if InitDefault has not been called yet.
func ExampleDefault() {
	db := sqlitevfs.Default()
	fmt.Println(db != nil)
}
