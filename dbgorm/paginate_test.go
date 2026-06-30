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
	"strings"
	"testing"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
	"github.com/phcp-tech/common-library-golang/dto"
)

// -----------------------------------------------------------------------
// Paginate
// -----------------------------------------------------------------------

func TestPaginate_UnlimitedMode(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck
	for i := range 5 {
		db.Create(&mockItem{Name: strings.Repeat("x", i+1)})
	}

	var results []mockItem
	db.Scopes(dbgorm.Paginate(1, -1)).Find(&results)
	if len(results) != 5 {
		t.Errorf("Paginate(1,-1) returned %d rows, want 5 (unlimited)", len(results))
	}
}

func TestPaginate_DefaultsWhenZeroLimit(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck

	var results []mockItem
	db.Scopes(dbgorm.Paginate(1, 0)).Find(&results)
	// limit=0 → defaultPageLimit; no panic
}

func TestPaginate_CapsAtMaxLimit(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck

	var results []mockItem
	db.Scopes(dbgorm.Paginate(1, 200)).Find(&results)
	// limit=200 > 100 → capped; no panic
}

func TestPaginate_DefaultPageWhenZero(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck
	db.Create(&mockItem{Name: "p"})

	var results []mockItem
	db.Scopes(dbgorm.Paginate(0, 1)).Find(&results)
	if len(results) != 1 {
		t.Errorf("Paginate(0,1) returned %d rows, want 1", len(results))
	}
}

// -----------------------------------------------------------------------
// OrderBy
// -----------------------------------------------------------------------

func TestOrderBy_UnknownColumn(t *testing.T) {
	db := openLocalDB(t)
	db.AutoMigrate(&mockItem{}) //nolint:errcheck
	db.Create(&mockItem{Name: "b"})
	db.Create(&mockItem{Name: "a"})

	allowed := map[string]string{"name": "name"}
	var results []mockItem
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
	db.Scopes(dbgorm.OrderBy(allowed, "name", "random")).Find(&results)
	if len(results) != 2 {
		t.Errorf("OrderBy random direction: got %d rows, want 2", len(results))
	}
	if results[0].Name != "a" {
		t.Errorf("OrderBy random direction: first = %q, want %q", results[0].Name, "a")
	}
}

// -----------------------------------------------------------------------
// SortSql
// isSafeSQLName and isSafeSQLIdentifierPath are private helpers called by
// SortSql; their branches are covered through the table-driven cases below.
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
		{
			name:          "unsafe sort falls back to id",
			para:          dto.PageParameter{Sort: "name; DROP TABLE users", Direction: "ASC"},
			wantSQL:       " ORDER BY id ASC",
			wantSort:      "id",
			wantDirection: "ASC",
		},
		{
			name:          "unsafe charset is ignored",
			para:          dto.PageParameter{Sort: "name", Direction: "ASC", Charset: "UTF8'); DROP TABLE users; --"},
			wantSQL:       " ORDER BY name ASC",
			wantSort:      "name",
			wantDirection: "ASC",
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

// -----------------------------------------------------------------------
// PageSql
// -----------------------------------------------------------------------

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
