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

package dbsqlx

import (
	"context"

	"github.com/vinovest/sqlx"
)

// Exec executes a statement with positional arguments and returns the number
// of affected rows.
func Exec(ctx context.Context, db sqlx.ExecerContext, query string, args ...any) (int64, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// NamedExec executes a named-parameter statement (e.g. "INSERT INTO t (name)
// VALUES (:name)") against arg — a struct or map[string]any providing the
// named values — and returns the number of affected rows.
func NamedExec(ctx context.Context, db sqlx.ExtContext, query string, arg any) (int64, error) {
	result, err := sqlx.NamedExecContext(ctx, db, query, arg)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Transact runs fn inside a transaction on db, committing on success and
// rolling back on error or panic (the panic is recovered just long enough to
// roll back, then re-raised to the caller).
//
// db must be *sqlx.DB for the outermost call. Nested calls made from inside fn
// — by passing the tx argument fn receives back into another Transact call
// with the same ctx — detect the ongoing transaction via ctx and reuse it
// instead of starting a new one; only the outermost call commits or rolls back.
//
// Do not call Transact with a *sqlx.Tx obtained outside of this ctx-propagation
// mechanism (e.g. via db.Beginx()) as the outermost db argument: the library
// only begins a transaction when db is *sqlx.DB, so a bare *sqlx.Tx passed
// directly causes a nil-pointer panic on commit/rollback.
func Transact(ctx context.Context, db sqlx.Queryable, fn func(ctx context.Context, tx sqlx.Queryable) error) error {
	return sqlx.TransactContext(ctx, db, fn)
}
