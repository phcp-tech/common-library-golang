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
	"os"
	"path/filepath"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dbgorm/sqlite"
)

// ExampleNewSQLite shows how to open an in-memory SQLite GORM database.
// In-memory databases are useful for tests and ephemeral data — all data
// is lost when the connection is closed.
func ExampleNewSQLite() {
	db, err := sqlite.NewSQLite(&sqlite.Config{
		Path: "file::memory:?cache=shared",
	})
	fmt.Println(err == nil)
	fmt.Println(db != nil)
	// Output:
	// true
	// true
}

// ExampleNewSQLite_file shows how to open a file-based SQLite database.
// SQLite creates the file automatically if it does not exist.
func ExampleNewSQLite_file() {
	path := filepath.Join(os.TempDir(), "dbgorm_example.db")
	defer os.Remove(path)

	db, err := sqlite.NewSQLite(&sqlite.Config{
		Path: "file:" + path + "?cache=shared",
	})
	fmt.Println(err == nil)
	fmt.Println(db != nil)
	// Output:
	// true
	// true
}

// ExampleInitDefault shows the default-instance pattern.
// Call InitDefault once at application startup; dbgorm.Default() returns
// the shared *gorm.DB for the process.
func ExampleInitDefault() {
	err := sqlite.InitDefault(&sqlite.Config{
		Path: "file::memory:?cache=shared",
	})
	fmt.Println(err == nil)
	fmt.Println(dbgorm.Default() != nil)
}
