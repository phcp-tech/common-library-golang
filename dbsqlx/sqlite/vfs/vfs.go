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

// Package vfs opens a SQLite database embedded inside a Go binary using
// [modernc.org/sqlite/vfs] and Go's [embed.FS].
//
// The embedded database file must reside at config/sqlite.db inside the
// provided embed.FS. This package is intended for read-heavy, binary-embedded
// datasets such as reference tables or look-up data shipped with the binary.
//
// Import this sub-package only when distributing a SQLite database as part of
// the binary. For regular file-based databases use [dbsqlx/sqlite] instead.
package vfs

import (
	"embed"
	"sync"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dbsqlx/sqlite"

	"github.com/vinovest/sqlx"
	"modernc.org/sqlite/vfs"
)

// New opens a sqlx database backed by a SQLite database embedded inside
// sqliteFS. The embedded file must be located at config/sqlite.db within
// the FS.
func New(sqliteFS *embed.FS) (*sqlx.DB, error) {
	fn, _, err := vfs.New(sqliteFS)
	if err != nil {
		return nil, err
	}
	return sqlite.NewSQLite(&sqlite.Config{
		Path: "file:config/sqlite.db?vfs=" + fn,
	})
}

var defaultOnce sync.Once

// InitDefault opens the embedded SQLite VFS database and stores it as the
// process-wide dbsqlx default. Only the first call has any effect; subsequent
// calls are a no-op (sync.Once).
func InitDefault(sqliteFS *embed.FS) error {
	var err error
	defaultOnce.Do(func() {
		var db *sqlx.DB
		db, err = New(sqliteFS)
		if err != nil {
			return
		}
		dbsqlx.SetDefault(db)
	})
	return err
}
