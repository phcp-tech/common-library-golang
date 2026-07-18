# CLAUDE.md

Guidance for Claude Code when working in this repository.

## Overview

`github.com/phcp-tech/common-library-golang` (Go 1.25) is the shared component library that ~15+ other repos in this workspace depend on for env/config loading, structured logging, database access, auth, HTTP client/server helpers, and application lifecycle orchestration (`bootstrap`). It ships three separate, non-interchangeable database layers (`dbgorm`, `dbsqlx`, `dbsqlc`) plus a `*/component` wrapper for nearly every package so it can be wired into `bootstrap.New()...Run()`. `README.md` is the primary reference — it has a full package table with links to per-package usage docs; treat this file as a supplement covering things that aren't obvious from reading that table.

## Commands

```bash
go mod tidy
go vet ./...
go build ./...
go test ./... -short   # skip integration tests that open a real TCP listener (e.g. httpserver)
go test ./...          # full run, including those integration tests
```

## Non-obvious things

- **Three DB layers, not one — pick carefully.** `dbgorm` is full GORM (has `AutoMigrate`, `migrate.go`). `dbsqlx` wraps `github.com/vinovest/sqlx` over `database/sql` (raw SQL, `SortSql`/`PageSql`/`Transact` helpers — this is what every sqlx-migrated service in this workspace imports). `dbsqlc` is neither: `dbsqlc/postgres.NewPostgres` returns a native `*pgxpool.Pool`, bypassing `database/sql` entirely. All three have same-named sibling packages per driver (`postgres`, `mysql`, `sqlite`, …) — importing `dbgorm/postgres` when you meant `dbsqlx/postgres` compiles fine and fails in a completely different way.
- **`bootstrap`'s component order is a hard convention, not a suggestion.** The first `Add()` call must be the env component and the second must be log — every other component reads config via `env.Env()` inside its own `Init()`, and `Init()` failures before `log.Init()` still get captured by Go's default `slog` handler writing to stderr. `Close()` runs in LIFO order across every phase (including `AddParallel` groups, which close concurrently as a unit), which is also why `log` is closed last: `env.Close()` is a no-op, so LIFO naturally makes the log line at shutdown the last thing written.
- **`dbsqlx.Default()`/`SetDefault()` is a single process-wide global**, not per-connection-pool state — guarded by an `RWMutex` but there's only ever one active `*sqlx.DB` per process. Calling `SetDefault` again (e.g. in a test that opens a second connection) silently replaces it for every other caller of `Default()` in the same process.
- **`dbsqlx.JSONRaw` deliberately isn't `encoding/json.RawMessage`.** Consumers built with `GOEXPERIMENT=jsonv2` find that `RawMessage` no longer satisfies `database/sql`'s generic `[]byte`-scan fallback (`Scan` errors for both NULL and non-NULL values), so `JSONRaw` implements `sql.Scanner`/`driver.Valuer` itself instead of relying on that fallback. It also isn't `gorm.io/datatypes.JSON` — that would pull the entire GORM module into every sqlx-only, no-ORM consumer that migrated away from GORM specifically to avoid it. `Value()` turns any zero-length `JSONRaw` into SQL `NULL` rather than an empty string, since Postgres `json`/`jsonb` columns reject `''` as invalid input (`SQLSTATE 22P02`).
- **The SQLite driver used by `dbsqlx/sqlite` can silently break on computed timestamp columns — and this isn't documented anywhere in the package itself.** `modernc.org/sqlite` reports a proper Go `time.Time`-shaped value only for a *bare* column reference; wrap that same column in any expression (`COALESCE(col, ?)`, a `CASE`, etc.) in a `SELECT`, and the driver returns an unparsed string instead — regardless of whether the underlying value is actually NULL. Scanning that into a `time.Time` (or a `*time.Time`) destination fails with `unsupported Scan, storing driver.Value type string into type *time.Time`. This only reproduces against the SQLite driver — Postgres handles the identical query correctly — so it only surfaces in this library's own SQLite-backed test helpers, not production. Several dependent repos hit this while trying to `COALESCE` a nullable timestamp column; the fix in each case was either to accept the column can't be tested via SQLite, or to select it bare and post-process NULL in Go rather than in SQL.
- **`SortSql`/`PageSql` (in `dbsqlx/paginate.go`) are the only sanctioned way to build `ORDER BY`/`LIMIT`/`OFFSET` from user input.** `SortSql` validates the sort column against `IsSafeSQLIdentifierPath` (falls back to `"id"` if unsafe) before interpolating it directly into the query string — there's no parameterized way to bind a column name, so this validation *is* the SQL-injection guard for that field, not a nicety.
- Tests that open a real TCP listener (`httpserver` integration tests) are **not** gated behind a build tag — they run unconditionally under plain `go test ./...` and only get skipped with `-short`, so a "fast" CI run must remember to pass that flag explicitly.
