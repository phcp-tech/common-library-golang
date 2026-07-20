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

package dbsqlx_test

import (
	"context"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dto"
)

// -----------------------------------------------------------------------
// SortSql
// IsSafeSQLName and IsSafeSQLIdentifierPath are helpers called by SortSql;
// their branches are covered through the table-driven cases below.
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
			name:          "unsafe sort falls back to id",
			para:          dto.PageParameter{Sort: "name; DROP TABLE users", Direction: "ASC"},
			wantSQL:       " ORDER BY id ASC",
			wantSort:      "id",
			wantDirection: "ASC",
		},
		{
			name:          "dotted path is safe",
			para:          dto.PageParameter{Sort: "users.name", Direction: "ASC"},
			wantSQL:       " ORDER BY users.name ASC",
			wantSort:      "users.name",
			wantDirection: "ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dbsqlx.SortSql(&tt.para)
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
			name:      "NoLimit disables pagination",
			para:      dto.PageParameter{Page: 2, Limit: dbsqlx.NoLimit},
			wantSQL:   "",
			wantPage:  2,
			wantLimit: dbsqlx.NoLimit,
		},
		{
			name:      "defaults invalid page and limit",
			para:      dto.PageParameter{},
			wantSQL:   " LIMIT 20 OFFSET 0",
			wantPage:  1,
			wantLimit: 20,
		},
		{
			// Regression test for the NoLimit/negative-Limit distinction:
			// only the exact NoLimit sentinel (-1) disables pagination — any
			// other negative value (e.g. from a malformed "?limit=-2" query
			// parameter, which strconv.Atoi parses without error) must NOT
			// also be treated as "no limit"; it falls back to
			// defaultPageLimit just like 0 does.
			name:      "negative limit other than NoLimit falls back to default, not unlimited",
			para:      dto.PageParameter{Page: 1, Limit: -2},
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
		{
			name:      "caps page at max",
			para:      dto.PageParameter{Page: 999999999, Limit: 20},
			wantSQL:   " LIMIT 20 OFFSET 1999980",
			wantPage:  100000,
			wantLimit: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dbsqlx.PageSql(&tt.para)
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

// TestSortSqlPageSql_ComposedForRawQuery shows the intended usage: append both
// clauses directly onto a base SELECT statement.
func TestSortSqlPageSql_ComposedForRawQuery(t *testing.T) {
	db := openTestDB(t)
	db.MustExec(`INSERT INTO users (name) VALUES (?)`, "carol")
	db.MustExec(`INSERT INTO users (name) VALUES (?)`, "alice")
	db.MustExec(`INSERT INTO users (name) VALUES (?)`, "bob")

	para := dto.PageParameter{Sort: "name", Direction: "ASC", Page: 1, Limit: 2}
	query := "SELECT * FROM users" + dbsqlx.SortSql(&para) + dbsqlx.PageSql(&para)

	var got []testUser
	if err := db.SelectContext(context.Background(), &got, query); err != nil {
		t.Fatalf("composed query error = %v", err)
	}
	if len(got) != 2 || got[0].Name != "alice" || got[1].Name != "bob" {
		t.Errorf("composed query: got %+v, want [alice, bob]", got)
	}
}
