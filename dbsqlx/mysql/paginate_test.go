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

package mysql_test

import (
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx/mysql"
	"github.com/phcp-tech/common-library-golang/dto"
)

func TestChineseSortSql(t *testing.T) {
	tests := []struct {
		name          string
		para          dto.PageParameter
		wantSQL       string
		wantSort      string
		wantDirection string
	}{
		{
			name:          "defaults to gbk ascending",
			para:          dto.PageParameter{Sort: "name"},
			wantSQL:       " ORDER BY CONVERT(name USING gbk) ASC",
			wantSort:      "name",
			wantDirection: "ASC",
		},
		{
			name:          "custom charset is lower-cased and descending direction",
			para:          dto.PageParameter{Sort: "name", Direction: " desc ", Charset: "GB18030"},
			wantSQL:       " ORDER BY CONVERT(name USING gb18030) DESC",
			wantSort:      "name",
			wantDirection: "DESC",
		},
		{
			name:          "invalid direction falls back to ascending",
			para:          dto.PageParameter{Sort: "name", Direction: "sideways"},
			wantSQL:       " ORDER BY CONVERT(name USING gbk) ASC",
			wantSort:      "name",
			wantDirection: "ASC",
		},
		{
			name:          "unsafe sort falls back to plain id sort",
			para:          dto.PageParameter{Sort: "name; DROP TABLE users", Direction: "ASC"},
			wantSQL:       " ORDER BY id ASC",
			wantSort:      "id",
			wantDirection: "ASC",
		},
		{
			name:          "unsafe charset falls back to plain column sort",
			para:          dto.PageParameter{Sort: "name", Direction: "ASC", Charset: "gbk); DROP TABLE users; --"},
			wantSQL:       " ORDER BY name ASC",
			wantSort:      "name",
			wantDirection: "ASC",
		},
		{
			name:          "empty sort falls back to plain id sort",
			para:          dto.PageParameter{},
			wantSQL:       " ORDER BY id ASC",
			wantSort:      "id",
			wantDirection: "ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mysql.ChineseSortSql(&tt.para)
			if got != tt.wantSQL {
				t.Fatalf("ChineseSortSql() = %q, want %q", got, tt.wantSQL)
			}
			if tt.para.Sort != tt.wantSort {
				t.Fatalf("ChineseSortSql() Sort = %q, want %q", tt.para.Sort, tt.wantSort)
			}
			if tt.para.Direction != tt.wantDirection {
				t.Fatalf("ChineseSortSql() Direction = %q, want %q", tt.para.Direction, tt.wantDirection)
			}
		})
	}
}
