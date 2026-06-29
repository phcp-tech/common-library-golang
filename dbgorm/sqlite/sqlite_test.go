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

package sqlite_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dbgorm/sqlite"
)

func TestDialectorRequiresPath(t *testing.T) {
	if _, err := sqlite.Dialector(&sqlite.Config{}); err != dbgorm.ErrMissingConfig {
		t.Fatalf("expected ErrMissingConfig, got %v", err)
	}
}

func TestOpenSQLiteInMemory(t *testing.T) {
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: "file::memory:?cache=shared"})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = dbgorm.Close(db) })

	if err := db.WithContext(context.Background()).Exec("select 1").Error; err != nil {
		t.Fatalf("select 1: %v", err)
	}
}

// TestNewSQLite_CreatesParentDirectory verifies that NewSQLite automatically
// creates the parent directory when the path contains a non-existent folder.
func TestNewSQLite_CreatesParentDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "level1", "level2")
	path := filepath.Join(dir, "app.db")

	db, err := sqlite.NewSQLite(&sqlite.Config{Path: path})
	if err != nil {
		t.Fatalf("NewSQLite with nested path: %v", err)
	}
	defer dbgorm.Close(db) //nolint:errcheck

	if _, statErr := os.Stat(path); statErr != nil {
		t.Errorf("expected db file at %s, got %v", path, statErr)
	}
}

// TestNewSQLite_ErrMissingPath covers the Dialector-error path in NewSQLite.
func TestNewSQLite_ErrMissingPath(t *testing.T) {
	_, err := sqlite.NewSQLite(&sqlite.Config{}) // empty path → ErrMissingConfig
	if !errors.Is(err, dbgorm.ErrMissingConfig) {
		t.Errorf("NewSQLite empty path: want ErrMissingConfig, got %v", err)
	}
}

// TestInitDefault_Success covers the success path of InitDefault using an in-memory DB.
func TestInitDefault_Success(t *testing.T) {
	prev := dbgorm.Default()
	t.Cleanup(func() { dbgorm.SetDefault(prev) })

	if err := sqlite.InitDefault(&sqlite.Config{Path: ":memory:"}); err != nil {
		t.Fatalf("InitDefault in-memory: %v", err)
	}
	if dbgorm.Default() == nil {
		t.Error("Default() is nil after successful InitDefault")
	}
}

// TestInitDefault_Error covers the error-return path of InitDefault.
func TestInitDefault_Error(t *testing.T) {
	err := sqlite.InitDefault(&sqlite.Config{}) // empty path → Dialector fails
	if !errors.Is(err, dbgorm.ErrMissingConfig) {
		t.Errorf("InitDefault empty path: want ErrMissingConfig, got %v", err)
	}
}

// TestAttach covers the Attach function using two in-memory SQLite databases.
func TestAttach(t *testing.T) {
	prev := dbgorm.Default()
	t.Cleanup(func() { dbgorm.SetDefault(prev) })

	// Open a primary in-memory DB and set it as default.
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	t.Cleanup(func() { dbgorm.Close(db) }) //nolint:errcheck
	dbgorm.SetDefault(db)

	// Attach a second in-memory DB under the alias "secondary".
	result := sqlite.Attach(":memory:", "secondary")
	if result.Error != nil {
		t.Errorf("Attach: %v", result.Error)
	}
}
