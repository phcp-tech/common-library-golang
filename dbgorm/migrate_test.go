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

package dbgorm_test

import (
	"context"
	"testing"

	dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
)

// migrateItem and migrateTag are local models used only by AutoMigrate tests.
type migrateItem struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Title string
}

type migrateTag struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string
}

// TestAutoMigrate_SingleModel verifies that AutoMigrate creates the table for a
// single model and returns nil.
func TestAutoMigrate_SingleModel(t *testing.T) {
	ctx := context.Background()
	db := openLocalDB(t)

	err := dbgorm.AutoMigrate(ctx, db, &migrateItem{})
	if err != nil {
		t.Fatalf("AutoMigrate single model: %v", err)
	}
	if !db.Migrator().HasTable(&migrateItem{}) {
		t.Error("expected table to exist after AutoMigrate")
	}
}

// TestAutoMigrate_MultipleModels verifies that AutoMigrate creates tables for
// all provided models in a single call.
func TestAutoMigrate_MultipleModels(t *testing.T) {
	ctx := context.Background()
	db := openLocalDB(t)

	err := dbgorm.AutoMigrate(ctx, db, &migrateItem{}, &migrateTag{})
	if err != nil {
		t.Fatalf("AutoMigrate multiple models: %v", err)
	}
	if !db.Migrator().HasTable(&migrateItem{}) {
		t.Error("expected migrateItem table to exist")
	}
	if !db.Migrator().HasTable(&migrateTag{}) {
		t.Error("expected migrateTag table to exist")
	}
}

// TestAutoMigrate_Idempotent verifies that calling AutoMigrate twice on the
// same model is safe and returns nil both times.
func TestAutoMigrate_Idempotent(t *testing.T) {
	ctx := context.Background()
	db := openLocalDB(t)

	if err := dbgorm.AutoMigrate(ctx, db, &migrateItem{}); err != nil {
		t.Fatalf("first AutoMigrate: %v", err)
	}
	if err := dbgorm.AutoMigrate(ctx, db, &migrateItem{}); err != nil {
		t.Fatalf("second AutoMigrate (idempotent): %v", err)
	}
}

// TestAutoMigrate_NoModels verifies that calling AutoMigrate with no models
// is a no-op and returns nil.
func TestAutoMigrate_NoModels(t *testing.T) {
	ctx := context.Background()
	db := openLocalDB(t)

	if err := dbgorm.AutoMigrate(ctx, db); err != nil {
		t.Errorf("AutoMigrate with no models: %v, want nil", err)
	}
}
