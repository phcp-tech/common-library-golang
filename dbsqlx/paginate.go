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

package dbsqlx

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/phcp-tech/common-library-golang/dto"
)

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
)

// SortSql builds an ORDER BY SQL clause from the sort fields in para.
// Defaults to ordering by "id ASC" when Sort is empty or unsafe.
// If para.Charset is set and safe, the sort column is wrapped with convert_to
// for locale-aware ordering.
func SortSql(para *dto.PageParameter) string {
	// default sort by Id
	if !isSafeSQLIdentifierPath(para.Sort) {
		para.Sort = "id"
	}

	// default direction is ASC, not errors if illegal
	para.Direction = strings.ToUpper(strings.TrimSpace(para.Direction))
	if para.Direction == "" {
		para.Direction = "ASC"
	} else if para.Direction != "ASC" && para.Direction != "DESC" {
		para.Direction = "ASC"
	}

	// order by charset
	var sqlstr string = ""
	if para.Charset != "" && isSafeSQLName(para.Charset) {
		// MySQL function is CONVERT
		//sqlstr += " ORDER BY CONVERT(" + para.Sort + " USING " + para.Charset + ") " + para.Direction
		sqlstr += " ORDER BY convert_to(" + para.Sort + ",'" + para.Charset + "') " + " " + para.Direction
	} else {
		sqlstr += " ORDER BY " + para.Sort + " " + para.Direction
	}
	return sqlstr
}

// PageSql builds a LIMIT/OFFSET SQL clause from the pagination fields in para.
// When Limit is -1, all records are returned without a LIMIT clause.
// Defaults to page 1 and defaultPageLimit when values are unset or invalid;
// caps Limit at maxPageLimit.
func PageSql(para *dto.PageParameter) string {
	var sqlstr string = ""
	if para.Limit == -1 {
		return sqlstr
	}

	if para.Page <= 0 {
		para.Page = 1
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

// isSafeSQLIdentifierPath reports whether value is a dot-separated path of
// safe SQL identifiers (e.g. "table.column"), guarding SortSql against
// injection via the Sort field.
func isSafeSQLIdentifierPath(value string) bool {
	if value == "" {
		return false
	}
	for _, part := range strings.Split(value, ".") {
		if !isSafeSQLName(part) {
			return false
		}
	}
	return true
}

// isSafeSQLName reports whether value is a safe single SQL identifier:
// letters, digits (not leading), and underscores only.
func isSafeSQLName(value string) bool {
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
