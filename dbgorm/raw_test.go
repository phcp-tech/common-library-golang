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
	"gorm.io/gorm"
)

// rawItem is a minimal model used only by raw SQL tests.
type rawItem struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	Name     string
	Category string
}

// openRawDB opens an in-memory SQLite database with rawItem migrated.
func openRawDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := openLocalDB(t)
	if err := db.AutoMigrate(&rawItem{}); err != nil {
		t.Fatalf("openRawDB migrate: %v", err)
	}
	return db
}

// -----------------------------------------------------------------------
// ExecRaw
// -----------------------------------------------------------------------

func TestExecRaw_UpdateReturnsRowsAffected(t *testing.T) {
	ctx := context.Background()
	db := openRawDB(t)
	db.Create(&rawItem{Name: "nail", Category: "fastener"})
	db.Create(&rawItem{Name: "bolt", Category: "fastener"})
	db.Create(&rawItem{Name: "brush", Category: "paint"})

	rows, err := dbgorm.ExecRaw(ctx, db,
		"UPDATE raw_items SET category = ? WHERE category = ?", "hardware", "fastener")
	if err != nil {
		t.Fatalf("ExecRaw error = %v", err)
	}
	if rows != 2 {
		t.Errorf("ExecRaw rows = %d, want 2", rows)
	}
}

func TestExecRaw_ZeroRowsAffected(t *testing.T) {
	ctx := context.Background()
	db := openRawDB(t)

	rows, err := dbgorm.ExecRaw(ctx, db,
		"UPDATE raw_items SET category = ? WHERE category = ?", "x", "nonexistent")
	if err != nil {
		t.Fatalf("ExecRaw zero rows: error = %v", err)
	}
	if rows != 0 {
		t.Errorf("ExecRaw zero rows: got %d, want 0", rows)
	}
}

// -----------------------------------------------------------------------
// ScanRaw
// -----------------------------------------------------------------------

func TestScanRaw_Scalar(t *testing.T) {
	ctx := context.Background()
	db := openRawDB(t)
	db.Create(&rawItem{Name: "a", Category: "tools"})
	db.Create(&rawItem{Name: "b", Category: "tools"})
	db.Create(&rawItem{Name: "c", Category: "other"})

	var count int64
	if err := dbgorm.ScanRaw(ctx, db, &count,
		"SELECT COUNT(*) FROM raw_items WHERE category = ?", "tools"); err != nil {
		t.Fatalf("ScanRaw scalar error = %v", err)
	}
	if count != 2 {
		t.Errorf("ScanRaw count = %d, want 2", count)
	}
}

func TestScanRaw_SingleRow(t *testing.T) {
	ctx := context.Background()
	db := openRawDB(t)
	db.Create(&rawItem{Name: "hammer", Category: "tools"})

	var item rawItem
	if err := dbgorm.ScanRaw(ctx, db, &item,
		"SELECT * FROM raw_items WHERE name = ?", "hammer"); err != nil {
		t.Fatalf("ScanRaw single row error = %v", err)
	}
	if item.Name != "hammer" {
		t.Errorf("ScanRaw single row: Name = %q, want %q", item.Name, "hammer")
	}
}

func TestScanRaw_MultipleRows(t *testing.T) {
	ctx := context.Background()
	db := openRawDB(t)
	db.Create(&rawItem{Name: "saw", Category: "tools"})
	db.Create(&rawItem{Name: "drill", Category: "tools"})
	db.Create(&rawItem{Name: "paint", Category: "other"})

	var items []rawItem
	if err := dbgorm.ScanRaw(ctx, db, &items,
		"SELECT * FROM raw_items WHERE category = ? ORDER BY name", "tools"); err != nil {
		t.Fatalf("ScanRaw multiple rows error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ScanRaw multiple rows: got %d, want 2", len(items))
	}
	if items[0].Name != "drill" || items[1].Name != "saw" {
		t.Errorf("ScanRaw order: got %v", items)
	}
}
