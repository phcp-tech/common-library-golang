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
	"os"
	"strings"

	"gorm.io/gorm"
)

// MigrateOptions controls AutoMigrate and optional mock data execution.
type MigrateOptions struct {
	Enabled    bool
	InsertMock bool
	MockFile   string
	Separator  string
}

// AutoMigrate migrates models and optionally executes mock SQL statements.
func AutoMigrate(ctx context.Context, db *gorm.DB, opts MigrateOptions, models ...any) error {
	if !opts.Enabled {
		return nil
	}
	for _, model := range models {
		if err := db.WithContext(ctx).AutoMigrate(model); err != nil {
			return err
		}
	}
	if !opts.InsertMock || opts.MockFile == "" {
		return nil
	}

	separator := opts.Separator
	if separator == "" {
		separator = "----"
	}

	content, err := os.ReadFile(opts.MockFile)
	if err != nil {
		return err
	}
	for _, sql := range strings.Split(string(content), separator) {
		sql = strings.TrimSpace(sql)
		if sql == "" || strings.HasPrefix(sql, "--") {
			continue
		}
		if _, err := ExecRaw(ctx, db, sql); err != nil {
			return err
		}
	}
	return nil
}
