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

// Create inserts record into the database and returns the created record.
// The returned pointer reflects auto-populated fields (primary key, timestamps, etc.).
func Create[T any](ctx context.Context, db *gorm.DB, record *T) (*T, error) {
	if err := db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

// FirstByID returns the first record of type T matching the primary key id.
func FirstByID[T any](ctx context.Context, db *gorm.DB, id any) (*T, error) {
	var model T
	if err := db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

// FirstWhere returns the first record of type T matching query and args.
func FirstWhere[T any](ctx context.Context, db *gorm.DB, query any, args ...any) (*T, error) {
	var model T
	if err := db.WithContext(ctx).Where(query, args...).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

// DeleteByID deletes the record of type T matching the primary key id.
func DeleteByID[T any](ctx context.Context, db *gorm.DB, id any) error {
	var model T
	tx := db.WithContext(ctx).Delete(&model, id)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteWhere deletes records for model matching query and args.
func DeleteWhere(ctx context.Context, db *gorm.DB, model any, query any, args ...any) error {
	return db.WithContext(ctx).Where(query, args...).Delete(model).Error
}

// UpdateByID updates columns of the record of type T matching the primary key id.
// values may be a struct (only non-zero fields are updated) or a map[string]any
// (all specified keys are updated regardless of zero value).
// Returns gorm.ErrRecordNotFound when no row is matched.
func UpdateByID[T any](ctx context.Context, db *gorm.DB, id any, values any) error {
	var model T
	tx := db.WithContext(ctx).Model(&model).Where("id = ?", id).Updates(values)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetCurrentId retrieves the maximum ID value from the table corresponding to model
// using GORM (COALESCE(MAX(id), 0)). model must be a pointer to a model struct
// (e.g., &Article{}). Returns the maximum ID and nil on success; returns 0 and nil
// for an empty table; returns 0 and an error on database failure.
func GetCurrentId[T any](ctx context.Context, db *gorm.DB) (uint64, error) {
	var model T
	var id uint64
	err := db.WithContext(ctx).Model(&model).Select("COALESCE(MAX(id), 0)").Scan(&id).Error
	if err != nil {
		return 0, err
	}
	return id, nil
}
