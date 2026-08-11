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

package mysql

import (
	"strings"

	"github.com/phcp-tech/common-library-golang/dbsqlx"
	"github.com/phcp-tech/common-library-golang/dto"
)

// defaultChineseCharset is used when para.Charset is empty.
const defaultChineseCharset = "gbk"

// ZhSortSql builds an ORDER BY clause that approximates pinyin ordering
// for a Chinese-text column, using MySQL's CONVERT(... USING charset) to
// re-encode the column before comparing byte order under that charset's
// default collation — gbk_chinese_ci sorts common Han characters roughly by
// pinyin, which is why this trick works. para.Charset selects the target
// charset (e.g. "gbk", "gb18030"); it defaults to "gbk" when empty.
//
// This is MySQL-specific: CONVERT(... USING ...) is not valid syntax on
// PostgreSQL, SQLite, or ClickHouse. dbsqlx.SortSql deliberately does not
// offer this — see its doc comment for why.
//
// para.Sort and the resolved charset are validated with the same
// identifier-safety rules dbsqlx.SortSql itself uses; if either is unsafe,
// ZhSortSql falls back to dbsqlx.SortSql's plain "ORDER BY <column>
// <direction>" behaviour instead (which also handles defaulting Sort to
// "id" and Direction to "ASC").
func ZhSortSql(para *dto.PageParameter) string {
	charset := strings.ToLower(para.Charset)
	if charset == "" {
		charset = defaultChineseCharset
	}
	if !dbsqlx.IsSafeSQLIdentifierPath(para.Sort) || !dbsqlx.IsSafeSQLName(charset) {
		return dbsqlx.SortSql(para)
	}

	para.Direction = dbsqlx.NormalizeSortDirection(para.Direction)
	return " ORDER BY CONVERT(" + para.Sort + " USING " + charset + ") " + para.Direction
}
