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
	"context"

	"gorm.io/gorm"
)

// ExecRaw executes a raw SQL statement with positional arguments.
func ExecRaw(ctx context.Context, db *gorm.DB, sql string, args ...any) (int64, error) {
	tx := db.WithContext(ctx).Exec(sql, args...)
	return tx.RowsAffected, tx.Error
}

// ScanRaw executes a raw SQL query and scans the result into dest.
func ScanRaw(ctx context.Context, db *gorm.DB, dest any, sql string, args ...any) error {
	return db.WithContext(ctx).Raw(sql, args...).Scan(dest).Error
}
