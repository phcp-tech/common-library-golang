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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dbgorm/sqlite"
	dbsqlite "github.com/phcp-tech/common-library-golang/dbgorm/sqlite"
	"github.com/phcp-tech/common-library-golang/health"

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

func TestFirstWhereDeleteWhereScopesAndRawSQL(t *testing.T) {
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

	rows, err := dbgorm.ExecRaw(ctx, db, "update test_users set status = ? where name = ?", "archived", "alice")
	if err != nil {
		t.Fatalf("exec raw: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected 1 affected row, got %d", rows)
	}

	var total int64
	if err := dbgorm.ScanRaw(ctx, db, &total, "select count(*) from test_users where status = ?", "archived"); err != nil {
		t.Fatalf("scan raw: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}

	if err := dbgorm.DeleteWhere(ctx, db, &testUser{}, "status = ?", "missing"); err != nil {
		t.Fatalf("delete where with zero rows should be nil: %v", err)
	}
}

func TestAutoMigrateHonorsEnabledFlag(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	type auditRecord struct {
		ID        uint `gorm:"primaryKey"`
		CreatedAt time.Time
	}

	if err := dbgorm.AutoMigrate(ctx, db, dbgorm.MigrateOptions{}, &auditRecord{}); err != nil {
		t.Fatalf("disabled migrate should be nil: %v", err)
	}
	if db.Migrator().HasTable(&auditRecord{}) {
		t.Fatalf("disabled migrate should not create table")
	}

	if err := dbgorm.AutoMigrate(ctx, db, dbgorm.MigrateOptions{Enabled: true}, &auditRecord{}); err != nil {
		t.Fatalf("enabled migrate: %v", err)
	}
	if !db.Migrator().HasTable(&auditRecord{}) {
		t.Fatalf("enabled migrate should create table")
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
// HealthChecker (moved from health_checker_test.go)
// -----------------------------------------------------------------------

func TestHealthChecker_ReturnsNonNil(t *testing.T) {
	if dbgorm.HealthChecker() == nil {
		t.Error("HealthChecker() returned nil")
	}
}

func TestHealthChecker_NoDefault(t *testing.T) {
	prev := dbgorm.Default()
	dbgorm.SetDefault(nil)
	defer dbgorm.SetDefault(prev)

	result := dbgorm.HealthChecker()(context.Background())
	if result.Name != "database" {
		t.Errorf("result.Name = %q, want %q", result.Name, "database")
	}
	if result.Status != health.StatusUnhealthy {
		t.Errorf("result.Status = %d, want StatusUnhealthy (%d)", result.Status, health.StatusUnhealthy)
	}
}

// TestHealthChecker_WithLiveDB covers the StatusHealthy path via a real SQLite DB.
func TestHealthChecker_WithLiveDB(t *testing.T) {
	prev := dbgorm.Default()
	dbgorm.SetDefault(openLocalDB(t))
	defer dbgorm.SetDefault(prev)

	result := dbgorm.HealthChecker()(context.Background())
	if result.Status != health.StatusHealthy {
		t.Errorf("result.Status = %d, want StatusHealthy (%d)", result.Status, health.StatusHealthy)
	}
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

// -----------------------------------------------------------------------
// Paginate — edge cases
// -----------------------------------------------------------------------

func TestPaginate_UnlimitedMode(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck
	for i := range 5 {
		db.Create(&mockItem{Name: strings.Repeat("x", i+1)})
	}

	var results []mockItem
	// limit=-1 disables pagination → all 5 rows returned
	db.Scopes(dbgorm.Paginate(1, -1)).Find(&results)
	if len(results) != 5 {
		t.Errorf("Paginate(1,-1) returned %d rows, want 5 (unlimited)", len(results))
	}
}

func TestPaginate_DefaultsWhenZeroLimit(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck

	// limit=0 → falls back to defaultPageLimit (20); no rows → empty but no panic
	var results []mockItem
	db.Scopes(dbgorm.Paginate(1, 0)).Find(&results)
	// verify the scope didn't panic
}

func TestPaginate_CapsAtMaxLimit(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck

	// limit=200 > maxPageLimit(100) → capped to 100; no panic
	var results []mockItem
	db.Scopes(dbgorm.Paginate(1, 200)).Find(&results)
}

func TestPaginate_DefaultPageWhenZero(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck
	db.Create(&mockItem{Name: "p"})

	var results []mockItem
	// page=0 → resets to page 1, limit=1 → returns the first row
	db.Scopes(dbgorm.Paginate(0, 1)).Find(&results)
	if len(results) != 1 {
		t.Errorf("Paginate(0,1) returned %d rows, want 1", len(results))
	}
}

// -----------------------------------------------------------------------
// OrderBy — edge cases
// -----------------------------------------------------------------------

func TestOrderBy_UnknownColumn(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck
	db.Create(&mockItem{Name: "b"})
	db.Create(&mockItem{Name: "a"})

	allowed := map[string]string{"name": "name"}
	var results []mockItem
	// "unknown" not in allowed → scope is a no-op, natural order preserved
	db.Scopes(dbgorm.OrderBy(allowed, "unknown", "ASC")).Find(&results)
	if len(results) != 2 {
		t.Errorf("OrderBy unknown column: got %d rows, want 2", len(results))
	}
}

func TestOrderBy_OtherDirection(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck
	db.Create(&mockItem{Name: "b"})
	db.Create(&mockItem{Name: "a"})

	allowed := map[string]string{"name": "name"}
	var results []mockItem
	// direction="random" is neither ASC nor DESC → treated as ASC (desc=false)
	db.Scopes(dbgorm.OrderBy(allowed, "name", "random")).Find(&results)
	if len(results) != 2 {
		t.Errorf("OrderBy random direction: got %d rows, want 2", len(results))
	}
	if results[0].Name != "a" {
		t.Errorf("OrderBy random direction: first = %q, want %q", results[0].Name, "a")
	}
}

// -----------------------------------------------------------------------
// AutoMigrate — all paths
// -----------------------------------------------------------------------

func TestAutoMigrate_Disabled(t *testing.T) {
	db := openLocalDB(t)
	err := dbgorm.AutoMigrate(context.Background(), db, dbgorm.MigrateOptions{Enabled: false}, &mockItem{})
	if err != nil {
		t.Errorf("AutoMigrate disabled: error = %v, want nil", err)
	}
	if db.Migrator().HasTable(&mockItem{}) {
		t.Error("AutoMigrate disabled: table must not be created")
	}
}

func TestAutoMigrate_NoMockFile(t *testing.T) {
	db := openLocalDB(t)
	// InsertMock=true but MockFile="" → no-op after migration
	err := dbgorm.AutoMigrate(context.Background(), db,
		dbgorm.MigrateOptions{Enabled: true, InsertMock: true, MockFile: ""},
		&mockItem{})
	if err != nil {
		t.Errorf("AutoMigrate no mock file: error = %v, want nil", err)
	}
}

func TestAutoMigrate_MockFileNotFound(t *testing.T) {
	db := openLocalDB(t)
	err := dbgorm.AutoMigrate(context.Background(), db,
		dbgorm.MigrateOptions{Enabled: true, InsertMock: true, MockFile: "nonexistent.sql"},
		&mockItem{})
	if err == nil {
		t.Error("AutoMigrate with missing mock file: want error, got nil")
	}
}

func TestAutoMigrate_WithMockSQL(t *testing.T) {
	db := openLocalDB(t)
	if err := db.AutoMigrate(&mockItem{}); err != nil {
		t.Fatalf("AutoMigrate model: %v", err)
	}

	// Write two INSERT statements separated by the default separator "----".
	mockSQL := "INSERT INTO mock_items (name) VALUES ('alpha')\n----\nINSERT INTO mock_items (name) VALUES ('beta')"
	f := filepath.Join(t.TempDir(), "mock.sql")
	if err := os.WriteFile(f, []byte(mockSQL), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	err := dbgorm.AutoMigrate(context.Background(), db,
		dbgorm.MigrateOptions{Enabled: true, InsertMock: true, MockFile: f},
		&mockItem{})
	if err != nil {
		t.Fatalf("AutoMigrate with mock SQL: %v", err)
	}

	var count int64
	db.Model(&mockItem{}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 rows after mock SQL, got %d", count)
	}
}

func TestAutoMigrate_CustomSeparator(t *testing.T) {
	db := openLocalDB(t)
	if err := db.AutoMigrate(&mockItem{}); err != nil {
		t.Fatalf("AutoMigrate model: %v", err)
	}

	mockSQL := "INSERT INTO mock_items (name) VALUES ('x');;;INSERT INTO mock_items (name) VALUES ('y')"
	f := filepath.Join(t.TempDir(), "mock.sql")
	if err := os.WriteFile(f, []byte(mockSQL), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	err := dbgorm.AutoMigrate(context.Background(), db,
		dbgorm.MigrateOptions{Enabled: true, InsertMock: true, MockFile: f, Separator: ";;;"},
		&mockItem{})
	if err != nil {
		t.Fatalf("AutoMigrate custom separator: %v", err)
	}

	var count int64
	db.Model(&mockItem{}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 rows after custom separator mock, got %d", count)
	}
}

func TestAutoMigrate_SkipsCommentsAndBlanks(t *testing.T) {
	db := openLocalDB(t)
	if err := db.AutoMigrate(&mockItem{}); err != nil {
		t.Fatalf("AutoMigrate model: %v", err)
	}

	// Comment segment is skipped; blank segment is skipped; only INSERT executes.
	mockSQL := "-- this is a comment\n----\n\n----\nINSERT INTO mock_items (name) VALUES ('only')"
	f := filepath.Join(t.TempDir(), "mock.sql")
	if err := os.WriteFile(f, []byte(mockSQL), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := dbgorm.AutoMigrate(context.Background(), db,
		dbgorm.MigrateOptions{Enabled: true, InsertMock: true, MockFile: f},
		&mockItem{}); err != nil {
		t.Fatalf("AutoMigrate skip comments: %v", err)
	}

	var count int64
	db.Model(&mockItem{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row (comment skipped), got %d", count)
	}
}
