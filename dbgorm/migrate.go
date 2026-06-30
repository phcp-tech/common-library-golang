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
	"log/slog"

	"gorm.io/gorm"
)

// AutoMigrate runs GORM auto-migration for all provided model types.
func AutoMigrate(ctx context.Context, db *gorm.DB, models ...any) error {
	for _, model := range models {
		if err := db.WithContext(ctx).AutoMigrate(model); err != nil {
			slog.Error("AutoMigrate failed", "model", model, "error", err)
			return err
		}
	}
	slog.Info("AutoMigrate completed successfully")
	return nil
}
