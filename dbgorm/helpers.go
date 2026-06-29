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
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

const (
	defaultPageLimit = 20
	maxPageLimit     = 100
)

// Paginate returns a GORM scope that applies limit and offset pagination.
func Paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if limit == -1 {
			return db
		}
		if page <= 0 {
			page = 1
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
