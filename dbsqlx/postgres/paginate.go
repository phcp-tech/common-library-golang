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

package postgres

import (
	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dto"
)

// defaultChineseCharset is used when para.Charset is empty.
const defaultChineseCharset = "GBK"

// ChineseSortSql builds an ORDER BY clause that approximates pinyin ordering
// for a Chinese-text column, using PostgreSQL's convert_to to re-encode the
// column before comparing byte order — GBK/GB18030 codepoints are assigned
// roughly in pinyin order for common Han characters, so sorting the
// converted bytes approximates a pinyin sort. para.Charset selects the
// target encoding (e.g. "GBK", "GB18030"); it defaults to "GBK" when empty.
//
// This is PostgreSQL-specific: convert_to does not exist on MySQL, SQLite,
// or ClickHouse. dbsqlx.SortSql deliberately does not offer this — see its
// doc comment for why.
//
// para.Sort and the resolved charset are validated with the same
// identifier-safety rules dbsqlx.SortSql itself uses; if either is unsafe,
// ChineseSortSql falls back to dbsqlx.SortSql's plain "ORDER BY <column>
// <direction>" behaviour instead (which also handles defaulting Sort to
// "id" and Direction to "ASC").
func ChineseSortSql(para *dto.PageParameter) string {
	charset := para.Charset
	if charset == "" {
		charset = defaultChineseCharset
	}
	if !dbsqlx.IsSafeSQLIdentifierPath(para.Sort) || !dbsqlx.IsSafeSQLName(charset) {
		return dbsqlx.SortSql(para)
	}

	para.Direction = dbsqlx.NormalizeSortDirection(para.Direction)
	return " ORDER BY convert_to(" + para.Sort + ",'" + charset + "') " + para.Direction
}
