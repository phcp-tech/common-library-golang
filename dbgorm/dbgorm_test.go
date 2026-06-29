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
	"github.com/phcp-tech/common-library-golang/dto"
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

func TestSortSql(t *testing.T) {
	tests := []struct {
		name          string
		para          dto.PageParameter
		wantSQL       string
		wantSort      string
		wantDirection string
	}{
		{
			name:          "defaults to id ascending",
			para:          dto.PageParameter{},
			wantSQL:       " ORDER BY id ASC",
			wantSort:      "id",
			wantDirection: "ASC",
		},
		{
			name:          "normalizes descending direction",
			para:          dto.PageParameter{Sort: "name", Direction: " desc "},
			wantSQL:       " ORDER BY name DESC",
			wantSort:      "name",
			wantDirection: "DESC",
		},
		{
			name:          "invalid direction falls back to ascending",
			para:          dto.PageParameter{Sort: "created_at", Direction: "random"},
			wantSQL:       " ORDER BY created_at ASC",
			wantSort:      "created_at",
			wantDirection: "ASC",
		},
		{
			name:          "charset wraps sort column",
			para:          dto.PageParameter{Sort: "name", Direction: "DESC", Charset: "UTF8"},
			wantSQL:       " ORDER BY convert_to(name,'UTF8')  DESC",
			wantSort:      "name",
			wantDirection: "DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dbgorm.SortSql(&tt.para)
			if got != tt.wantSQL {
				t.Fatalf("SortSql() = %q, want %q", got, tt.wantSQL)
			}
			if tt.para.Sort != tt.wantSort {
				t.Fatalf("SortSql() Sort = %q, want %q", tt.para.Sort, tt.wantSort)
			}
			if tt.para.Direction != tt.wantDirection {
				t.Fatalf("SortSql() Direction = %q, want %q", tt.para.Direction, tt.wantDirection)
			}
		})
	}
}

func TestPageSql(t *testing.T) {
	tests := []struct {
		name      string
		para      dto.PageParameter
		wantSQL   string
		wantPage  int
		wantLimit int
	}{
		{
			name:      "calculates limit and offset",
			para:      dto.PageParameter{Page: 3, Limit: 25},
			wantSQL:   " LIMIT 25 OFFSET 50",
			wantPage:  3,
			wantLimit: 25,
		},
		{
			name:      "limit minus one disables pagination",
			para:      dto.PageParameter{Page: 2, Limit: -1},
			wantSQL:   "",
			wantPage:  2,
			wantLimit: -1,
		},
		{
			name:      "defaults invalid page and limit",
			para:      dto.PageParameter{},
			wantSQL:   " LIMIT 20 OFFSET 0",
			wantPage:  1,
			wantLimit: 20,
		},
		{
			name:      "caps limit at max",
			para:      dto.PageParameter{Page: 2, Limit: 200},
			wantSQL:   " LIMIT 100 OFFSET 100",
			wantPage:  2,
			wantLimit: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dbgorm.PageSql(&tt.para)
			if got != tt.wantSQL {
				t.Fatalf("PageSql() = %q, want %q", got, tt.wantSQL)
			}
			if tt.para.Page != tt.wantPage {
				t.Fatalf("PageSql() Page = %d, want %d", tt.para.Page, tt.wantPage)
			}
			if tt.para.Limit != tt.wantLimit {
				t.Fatalf("PageSql() Limit = %d, want %d", tt.para.Limit, tt.wantLimit)
			}
		})
	}
}

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
