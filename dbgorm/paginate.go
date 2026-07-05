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

package dbgorm

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/phcp-tech/common-library-golang/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
	// maxPage bounds how deep OFFSET-based pagination can go. Page itself
	// can't carry a SQL-injection payload (it's an int, not a string — see
	// PageSql's doc comment), but an unbounded Page still lets a caller
	// request an enormous OFFSET, which databases must still count past
	// even though it returns no rows — a cheap way to force expensive scans.
	maxPage = 100000
)

// Paginate returns a GORM scope that applies limit and offset pagination.
// Caps limit at maxPageLimit and page at maxPage — see PageSql's doc comment
// for why an unbounded page is a resource-usage guard, not a security fix.
func Paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if limit == -1 {
			return db
		}
		if page <= 0 {
			page = 1
		}
		if page > maxPage {
			page = maxPage
		}
		if limit <= 0 {
			limit = defaultPageLimit
		}
		if limit > maxPageLimit {
			limit = maxPageLimit
		}
		return db.Offset((page - 1) * limit).Limit(limit)
	}
}

// OrderBy returns a GORM scope that orders by an allow-listed column.
func OrderBy(allowed map[string]string, sort, direction string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		column, ok := allowed[sort]
		if !ok || column == "" {
			return db
		}

		desc := strings.EqualFold(strings.TrimSpace(direction), "DESC")
		if !desc && !strings.EqualFold(strings.TrimSpace(direction), "ASC") {
			desc = false
		}

		return db.Order(clause.OrderByColumn{
			Column: clause.Column{Name: column},
			Desc:   desc,
		})
	}
}

// SortSql builds an ORDER BY SQL clause from the sort fields in para.
// Defaults to ordering by "id ASC" when Sort is empty or unsafe.
//
// SortSql only ever produces a plain "ORDER BY <column> <direction>" clause —
// it deliberately has no locale/charset-aware sorting, since that requires a
// database-specific SQL function (e.g. PostgreSQL's convert_to, MySQL's
// CONVERT(... USING ...)) that this dialect-agnostic root package cannot
// pick correctly on its own. For approximate pinyin ordering of Chinese
// text, see the dedicated ChineseSortSql in the relevant dialect package
// (dbgorm/postgres, dbgorm/mysql) instead.
func SortSql(para *dto.PageParameter) string {
	// default sort by Id
	if !IsSafeSQLIdentifierPath(para.Sort) {
		para.Sort = "id"
	}
	para.Direction = NormalizeSortDirection(para.Direction)
	return " ORDER BY " + para.Sort + " " + para.Direction
}

// NormalizeSortDirection returns direction upper-cased and trimmed, falling
// back to "ASC" when it is empty or not one of "ASC"/"DESC". Exported so
// dialect packages building their own ORDER BY expressions (e.g.
// ChineseSortSql in dbgorm/postgres, dbgorm/mysql) apply the exact same
// direction rules SortSql does.
func NormalizeSortDirection(direction string) string {
	direction = strings.ToUpper(strings.TrimSpace(direction))
	if direction != "ASC" && direction != "DESC" {
		return "ASC"
	}
	return direction
}

// PageSql builds a LIMIT/OFFSET SQL clause from the pagination fields in para.
// When Limit is -1, all records are returned without a LIMIT clause.
// Defaults to page 1 and defaultPageLimit when values are unset or invalid;
// caps Limit at maxPageLimit and Page at maxPage.
//
// Page and Limit are ints, not strings, so — unlike Sort — they carry no SQL
// injection risk: strconv.Itoa on an int can only ever produce a plain
// decimal digit string. The Page cap here is a resource-usage guard instead,
// not a security fix: without it, an arbitrarily large Page still produces
// a syntactically valid but enormous OFFSET that the database must count
// past before returning zero rows.
func PageSql(para *dto.PageParameter) string {
	var sqlstr string = ""
	if para.Limit == -1 {
		return sqlstr
	}

	if para.Page <= 0 {
		para.Page = 1
	}
	if para.Page > maxPage {
		para.Page = maxPage
	}
	if para.Limit <= 0 {
		para.Limit = defaultPageLimit
	}
	if para.Limit > maxPageLimit {
		para.Limit = maxPageLimit
	}

	if para.Page >= 1 && para.Limit >= 1 {
		sqlstr += " LIMIT " + strconv.Itoa(para.Limit) + " OFFSET " + strconv.Itoa((para.Page-1)*para.Limit)
	}

	return sqlstr
}

// IsSafeSQLIdentifierPath reports whether value is a dot-separated path of
// safe SQL identifiers (e.g. "table.column"), guarding SortSql against
// injection via the Sort field. Exported so dialect packages building their
// own ORDER BY expressions (e.g. ChineseSortSql in dbgorm/postgres,
// dbgorm/mysql) can validate a column name with the same rules.
func IsSafeSQLIdentifierPath(value string) bool {
	if value == "" {
		return false
	}
	if !strings.Contains(value, ".") {
		return IsSafeSQLName(value)
	}
	for _, part := range strings.Split(value, ".") {
		if !IsSafeSQLName(part) {
			return false
		}
	}
	return true
}

// IsSafeSQLName reports whether value is a safe single SQL identifier:
// letters, digits (not leading), and underscores only. Also usable to
// validate a charset/encoding name interpolated directly into SQL (e.g. by
// ChineseSortSql), since it follows the same safe-token shape.
func IsSafeSQLName(value string) bool {
	if value == "" {
		return false
	}
	for i, r := range value {
		if r == '_' || unicode.IsLetter(r) || (i > 0 && unicode.IsDigit(r)) {
			continue
		}
		return false
	}
	return true
}
