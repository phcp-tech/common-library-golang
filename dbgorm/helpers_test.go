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
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dbgorm/sqlite"
	dbsqlite "github.com/phcp-tech/common-library-golang/dbgorm/sqlite"

	"gorm.io/gorm"
)

type testUser struct {
	ID     uint `gorm:"primaryKey"`
	Name   string
	Status string
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := dbsqlite.NewSQLite(&dbsqlite.Config{
		Path: "file::memory:?cache=shared",
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = dbgorm.Close(db) })

	if err := db.AutoMigrate(&testUser{}); err != nil {
		t.Fatalf("migrate test user: %v", err)
	}
	return db
}

func TestOpenRejectsNilDialector(t *testing.T) {
	_, err := dbgorm.Open(nil, nil)
	if !errors.Is(err, dbgorm.ErrNilDialector) {
		t.Fatalf("expected ErrNilDialector, got %v", err)
	}
}

func TestOpenUsesDefaultSlogLogger(t *testing.T) {
	var buf bytes.Buffer
	previousDefaultLogger := slog.Default()
	configuredLogger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(configuredLogger)
	t.Cleanup(func() { slog.SetDefault(previousDefaultLogger) })

	db, err := dbsqlite.NewSQLite(
		&dbsqlite.Config{
			Path: "file::memory:?cache=shared",
		},
	)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = dbgorm.Close(db) })

	if err := db.Exec("select * from missing_config_log_table").Error; err == nil {
		t.Fatalf("expected invalid query to fail")
	}
	if !strings.Contains(buf.String(), "missing_config_log_table") {
		t.Fatalf("expected custom log to receive gorm SQL log, got %q", buf.String())
	}
	t.Logf("default slog output:\n%s", strings.TrimSpace(buf.String()))
}

func TestDefaultLifecycle(t *testing.T) {
	dbgorm.SetDefault(nil)
	if got := dbgorm.Default(); got != nil {
		t.Fatalf("expected nil default before initialization, got %v", got)
	}

	db := openTestDB(t)
	dbgorm.SetDefault(db)
	t.Cleanup(func() { dbgorm.SetDefault(nil) })

	if got := dbgorm.Default(); got != db {
		t.Fatalf("default db mismatch")
	}
}

func TestFirstByIDAndDeleteByIDUseGormNotFound(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	created := testUser{Name: "alice", Status: "active"}
	if err := db.WithContext(ctx).Create(&created).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	found, err := dbgorm.FirstByID[testUser](ctx, db, created.ID)
	if err != nil {
		t.Fatalf("first by id: %v", err)
	}
	if found.Name != "alice" {
		t.Fatalf("expected alice, got %q", found.Name)
	}

	if _, err := dbgorm.FirstByID[testUser](ctx, db, 9999); !dbgorm.IsNotFound(err) {
		t.Fatalf("expected not found from FirstByID, got %v", err)
	}

	if err := dbgorm.DeleteByID[testUser](ctx, db, created.ID); err != nil {
		t.Fatalf("delete by id: %v", err)
	}
	if err := dbgorm.DeleteByID[testUser](ctx, db, 9999); !dbgorm.IsNotFound(err) {
		t.Fatalf("expected not found from DeleteByID, got %v", err)
	}
}

func TestFirstWhereDeleteWhereAndScopes(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	users := []testUser{
		{Name: "alice", Status: "active"},
		{Name: "bob", Status: "inactive"},
		{Name: "carol", Status: "active"},
	}
	if err := db.WithContext(ctx).Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}

	found, err := dbgorm.FirstWhere[testUser](ctx, db, "status = ?", "inactive")
	if err != nil {
		t.Fatalf("first where: %v", err)
	}
	if found.Name != "bob" {
		t.Fatalf("expected bob, got %q", found.Name)
	}

	var ordered []testUser
	allowed := map[string]string{"name": "name"}
	if err := db.WithContext(ctx).
		Scopes(dbgorm.OrderBy(allowed, "name", "DESC"), dbgorm.Paginate(1, 2)).
		Find(&ordered).Error; err != nil {
		t.Fatalf("query ordered users: %v", err)
	}
	if len(ordered) != 2 || ordered[0].Name != "carol" || ordered[1].Name != "bob" {
		t.Fatalf("unexpected ordered page: %#v", ordered)
	}

	if err := dbgorm.DeleteWhere(ctx, db, &testUser{}, "status = ?", "missing"); err != nil {
		t.Fatalf("delete where with zero rows should be nil: %v", err)
	}
}

func TestCreate_Success(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	created, err := dbgorm.Create(ctx, db, &testUser{Name: "bob", Status: "active"})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if created.ID == 0 {
		t.Error("Create: expected auto-populated ID, got 0")
	}
	if created.Name != "bob" {
		t.Errorf("Create: Name = %q, want %q", created.Name, "bob")
	}
}

func TestUpdateByID_Success(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	created := testUser{Name: "alice", Status: "active"}
	db.WithContext(ctx).Create(&created) //nolint:errcheck

	err := dbgorm.UpdateByID[testUser](ctx, db, created.ID, map[string]any{"status": "inactive"})
	if err != nil {
		t.Fatalf("UpdateByID error = %v", err)
	}

	updated, _ := dbgorm.FirstByID[testUser](ctx, db, created.ID)
	if updated.Status != "inactive" {
		t.Errorf("Status = %q, want %q", updated.Status, "inactive")
	}
}

func TestUpdateByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	err := dbgorm.UpdateByID[testUser](ctx, db, 9999, map[string]any{"status": "x"})
	if !dbgorm.IsNotFound(err) {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestUpdateWhere_UpdatesMatchingRows(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	db.WithContext(ctx).Create(&testUser{Name: "alice", Status: "active"})   //nolint:errcheck
	db.WithContext(ctx).Create(&testUser{Name: "bob", Status: "active"})     //nolint:errcheck
	db.WithContext(ctx).Create(&testUser{Name: "carol", Status: "inactive"}) //nolint:errcheck

	err := dbgorm.UpdateWhere(ctx, db, &testUser{}, map[string]any{"status": "archived"}, "status = ?", "active")
	if err != nil {
		t.Fatalf("UpdateWhere error = %v", err)
	}

	var count int64
	db.WithContext(ctx).Model(&testUser{}).Where("status = ?", "archived").Count(&count)
	if count != 2 {
		t.Errorf("expected 2 archived users, got %d", count)
	}
}

func TestUpdateWhere_ZeroMatchIsSuccess(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// No rows match — UpdateWhere treats zero affected rows as success.
	err := dbgorm.UpdateWhere(ctx, db, &testUser{}, map[string]any{"status": "x"}, "status = ?", "nonexistent")
	if err != nil {
		t.Errorf("UpdateWhere with no matching rows should return nil, got %v", err)
	}
}

// -----------------------------------------------------------------------
// mockItem — minimal GORM model used by the tests below.
// -----------------------------------------------------------------------

type mockItem struct {
	ID   uint `gorm:"primaryKey;autoIncrement"`
	Name string
}

// openLocalDB opens a private in-memory SQLite using the LOCAL dbgorm/sqlite package.
func openLocalDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := sqlite.NewSQLite(&sqlite.Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("openLocalDB: %v", err)
	}
	t.Cleanup(func() { dbgorm.Close(db) }) //nolint:errcheck
	return db
}


// -----------------------------------------------------------------------
// Open / Close
// -----------------------------------------------------------------------

func TestOpen_NilDialector(t *testing.T) {
	_, err := dbgorm.Open(nil, &dbgorm.GormConfig{})
	if !errors.Is(err, dbgorm.ErrNilDialector) {
		t.Errorf("Open(nil) error = %v, want ErrNilDialector", err)
	}
}

func TestOpen_WithCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	dialector, _ := sqlite.Dialector(&sqlite.Config{Path: ":memory:"})
	db, err := dbgorm.Open(dialector, &dbgorm.GormConfig{
		MaxOpenConns: 1,
		Logger:       logger,
	})
	if err != nil {
		t.Fatalf("Open with custom logger error = %v", err)
	}
	dbgorm.Close(db) //nolint:errcheck
}

func TestClose_NilDB(t *testing.T) {
	if err := dbgorm.Close(nil); err != nil {
		t.Errorf("Close(nil) = %v, want nil", err)
	}
}
