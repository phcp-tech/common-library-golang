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

// Package dbsqlx_test uses the pure-Go SQLite driver (modernc.org/sqlite) as a
// dependency-light vehicle for exercising the driver-agnostic root package —
// it does not depend on the dbsqlx/sqlite adapter package.
package dbsqlx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/phcp-tech/common-library-golang/dbsqlx"

	"github.com/vinovest/sqlx"
	_ "modernc.org/sqlite"
)

// testUser is the shared fixture model used across dbsqlx_test files.
type testUser struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Status string `db:"status"`
}

// openTestDB opens a private in-memory SQLite database via dbsqlx.Open and
// creates the users table used by the tests in this package.
func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := dbsqlx.Open("sqlite", ":memory:", nil)
	if err != nil {
		t.Fatalf("openTestDB: %v", err)
	}
	t.Cleanup(func() { _ = dbsqlx.Close(db) })

	if _, err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, status TEXT NOT NULL DEFAULT '')`); err != nil {
		t.Fatalf("openTestDB create table: %v", err)
	}
	return db
}

// -----------------------------------------------------------------------
// Exec
// -----------------------------------------------------------------------

func TestExec_RowsAffected(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	db.MustExec(`INSERT INTO users (name, status) VALUES (?, ?)`, "alice", "active")
	db.MustExec(`INSERT INTO users (name, status) VALUES (?, ?)`, "bob", "active")

	rows, err := dbsqlx.Exec(ctx, db, `UPDATE users SET status = ? WHERE status = ?`, "archived", "active")
	if err != nil {
		t.Fatalf("Exec error = %v", err)
	}
	if rows != 2 {
		t.Errorf("Exec rows = %d, want 2", rows)
	}
}

func TestExec_ZeroRowsAffected(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	rows, err := dbsqlx.Exec(ctx, db, `UPDATE users SET status = ? WHERE id = ?`, "x", 9999)
	if err != nil {
		t.Fatalf("Exec zero rows: error = %v, want nil", err)
	}
	if rows != 0 {
		t.Errorf("Exec zero rows: got %d, want 0", rows)
	}
}

func TestExec_InvalidSQL(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	_, err := dbsqlx.Exec(ctx, db, `NOT VALID SQL`)
	if err == nil {
		t.Error("Exec with invalid SQL: want error, got nil")
	}
}

// -----------------------------------------------------------------------
// NamedExec
// -----------------------------------------------------------------------

func TestNamedExec_WithMap(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	rows, err := dbsqlx.NamedExec(ctx, db, `INSERT INTO users (name, status) VALUES (:name, :status)`,
		map[string]any{"name": "carol", "status": "active"})
	if err != nil {
		t.Fatalf("NamedExec error = %v", err)
	}
	if rows != 1 {
		t.Errorf("NamedExec rows = %d, want 1", rows)
	}

	var got testUser
	if err := db.GetContext(ctx, &got, `SELECT * FROM users WHERE name = ?`, "carol"); err != nil {
		t.Fatalf("verify insert: %v", err)
	}
	if got.Status != "active" {
		t.Errorf("Status = %q, want %q", got.Status, "active")
	}
}

func TestNamedExec_WithStruct(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	arg := testUser{Name: "dave", Status: "active"}
	rows, err := dbsqlx.NamedExec(ctx, db, `INSERT INTO users (name, status) VALUES (:name, :status)`, arg)
	if err != nil {
		t.Fatalf("NamedExec with struct error = %v", err)
	}
	if rows != 1 {
		t.Errorf("NamedExec rows = %d, want 1", rows)
	}
}

func TestNamedExec_MissingField(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	_, err := dbsqlx.NamedExec(ctx, db, `INSERT INTO users (name) VALUES (:missing_field)`,
		map[string]any{"name": "z"})
	if err == nil {
		t.Error("NamedExec with a field absent from arg: want error, got nil")
	}
}

// -----------------------------------------------------------------------
// Transact
// -----------------------------------------------------------------------

func TestTransact_CommitsOnSuccess(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	err := dbsqlx.Transact(ctx, db, func(ctx context.Context, tx sqlx.Queryable) error {
		_, err := dbsqlx.Exec(ctx, tx, `INSERT INTO users (name) VALUES (?)`, "alice")
		return err
	})
	if err != nil {
		t.Fatalf("Transact error = %v", err)
	}

	var count int64
	if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users`); err != nil {
		t.Fatalf("verify count: %v", err)
	}
	if count != 1 {
		t.Errorf("count after commit = %d, want 1", count)
	}
}

func TestTransact_RollsBackOnError(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	db.MustExec(`INSERT INTO users (name) VALUES (?)`, "existing")

	wantErr := errors.New("forced failure")
	err := dbsqlx.Transact(ctx, db, func(ctx context.Context, tx sqlx.Queryable) error {
		if _, err := dbsqlx.Exec(ctx, tx, `INSERT INTO users (name) VALUES (?)`, "should-not-persist"); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("Transact error = %v, want %v", err, wantErr)
	}

	var count int64
	if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users`); err != nil {
		t.Fatalf("verify count: %v", err)
	}
	if count != 1 {
		t.Errorf("count after rollback = %d, want 1 (insert must not persist)", count)
	}
}

func TestTransact_RollsBackOnPanic(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic to propagate out of Transact")
			}
		}()
		_ = dbsqlx.Transact(ctx, db, func(ctx context.Context, tx sqlx.Queryable) error {
			_, _ = dbsqlx.Exec(ctx, tx, `INSERT INTO users (name) VALUES (?)`, "should-not-persist")
			panic("boom")
		})
	}()

	var count int64
	if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users`); err != nil {
		t.Fatalf("verify count: %v", err)
	}
	if count != 0 {
		t.Errorf("count after panic rollback = %d, want 0", count)
	}
}

func TestTransact_NestedReusesOuterTransaction(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	err := dbsqlx.Transact(ctx, db, func(ctx context.Context, tx sqlx.Queryable) error {
		if _, err := dbsqlx.Exec(ctx, tx, `INSERT INTO users (name) VALUES (?)`, "outer"); err != nil {
			return err
		}
		// Nested call: pass the same ctx so the inner Transact detects the
		// ongoing transaction and reuses it instead of beginning a new one.
		return dbsqlx.Transact(ctx, tx, func(ctx context.Context, inner sqlx.Queryable) error {
			_, err := dbsqlx.Exec(ctx, inner, `INSERT INTO users (name) VALUES (?)`, "inner")
			return err
		})
	})
	if err != nil {
		t.Fatalf("nested Transact error = %v", err)
	}

	var count int64
	if err := db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users`); err != nil {
		t.Fatalf("verify count: %v", err)
	}
	if count != 2 {
		t.Errorf("count after nested commit = %d, want 2", count)
	}
}
