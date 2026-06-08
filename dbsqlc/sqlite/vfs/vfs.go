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

// Package vfs opens a SQLite database embedded in a Go binary via
// modernc.org/sqlite/vfs. Import this package only when the SQLite database
// file is distributed as part of the binary (embed.FS); for regular file-based
// databases use the parent dbsqlc/sqlite package instead.
package vfs

import (
	"database/sql"
	"embed"

	"github.com/phcp-tech/common-library-golang/dbsqlc/sqlite"
	sqlitevfs "modernc.org/sqlite/vfs"
)

// New mounts the provided embedded filesystem as a SQLite VFS and opens
// a *sql.DB connection to the embedded database file at "config/sqlite.db".
// The embedded FS is registered with modernc.org/sqlite/vfs and the resulting
// VFS name is used to construct the SQLite URI. Returns the open *sql.DB or an
// error if VFS registration or database opening fails.
func New(sqliteFS *embed.FS) (*sql.DB, error) {
	//mount vfs to read sqlite DB
	fn, _, err := sqlitevfs.New(sqliteFS)
	if err != nil {
		return nil, err
	}

	// TODO: db file must be config/sqlite.db
	conf := sqlite.Config{
		Path: "file:config/sqlite.db?vfs=" + fn,
	}
	return sqlite.NewSQLite(&conf)
}
