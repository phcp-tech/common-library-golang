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
	"embed"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	sqlitevfs "github.com/phcp-tech/common-library-golang/dbsqlx/sqlite/vfs"
)

// testFS embeds config/sqlite.db so New can mount it as a VFS.
//
//go:embed config
var testFS embed.FS

func TestNew_OpensEmbeddedDB(t *testing.T) {
	db, err := sqlitevfs.New(&testFS)
	if err != nil {
		t.Fatalf("New error = %v", err)
	}
	if db == nil {
		t.Fatal("New returned nil")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("Ping() error = %v, want nil", err)
	}
}

func TestNew_CanQuery(t *testing.T) {
	db, err := sqlitevfs.New(&testFS)
	if err != nil {
		t.Fatalf("New error = %v", err)
	}
	defer db.Close()

	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM meta").Scan(&n); err != nil {
		t.Fatalf("QueryRow error = %v", err)
	}
	if n < 0 {
		t.Errorf("COUNT(*) = %d, want >= 0", n)
	}
}

func TestSingleton_Lifecycle(t *testing.T) {
	initial := dbsqlx.Default()

	if err := sqlitevfs.InitDefault(&testFS); err != nil {
		t.Fatalf("InitDefault error = %v", err)
	}

	if dbsqlx.Default() == nil {
		t.Fatal("Default() is nil after InitDefault")
	}

	// Second call is a no-op (sync.Once).
	first := dbsqlx.Default()
	if err := sqlitevfs.InitDefault(&testFS); err != nil {
		t.Errorf("second InitDefault should return nil; got %v", err)
	}
	if dbsqlx.Default() != first {
		t.Error("second InitDefault must not replace the existing instance")
	}

	_ = initial
}
