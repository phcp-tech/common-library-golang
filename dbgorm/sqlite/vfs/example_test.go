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

package vfs_test

import (
	"fmt"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	sqlitevfs "github.com/phcp-tech/common-library-golang/dbgorm/sqlite/vfs"
)

// ExampleNew shows how to open a GORM database backed by an embedded SQLite file.
// The embedded FS must contain config/sqlite.db.
//
//	//go:embed config
//	var sqliteFS embed.FS
//
// testFS is declared in vfs_test.go and shared across the test package.
func ExampleNew() {
	db, err := sqlitevfs.New(&testFS)
	fmt.Println(err == nil)
	fmt.Println(db != nil)
	// Output:
	// true
	// true
}

// ExampleInitDefault shows the singleton pattern for an embedded SQLite VFS database.
// Call InitDefault once at application startup; dbgorm.Default() returns the
// shared *gorm.DB. Subsequent calls are silently ignored (sync.Once).
func ExampleInitDefault() {
	// InitDefault is a no-op here because TestSingleton_Lifecycle ran first.
	_ = sqlitevfs.InitDefault(&testFS)
	fmt.Println(dbgorm.Default() != nil)
}
