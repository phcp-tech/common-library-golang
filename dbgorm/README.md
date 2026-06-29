# dbgorm

`dbgorm` is a thin wrapper around GORM. It does not replace GORM, hide
`*gorm.DB`, or provide a repository framework. Its main purpose is to make
database initialization consistent while keeping normal database operations in
plain GORM.

## Features

- Common GORM configuration and connection pool setup.
- Optional process-wide default database instance.
- Driver adapters for MySQL, PostgreSQL, and SQLite.
- Small helpers for common read/delete/not-found patterns.
- GORM scopes for pagination and allow-listed sorting.
- Raw SQL helpers with parameter binding.
- Optional migration helper.

## Package Layout

The root package does not import concrete database drivers:

```go
import dbgorm "github.com/phcp-tech/common-library-golang/dbgorm"
```

Concrete drivers live in adapter packages:

```go
import "github.com/phcp-tech/common-library-golang/dbgorm/mysql"
import "github.com/phcp-tech/common-library-golang/dbgorm/postgres"
import "github.com/phcp-tech/common-library-golang/dbgorm/sqlite"
```

This keeps the root package independent from driver imports. Applications should
only import the adapter packages they actually use.

The adapter packages expose flat config structs for application code. The root
package only keeps the lower-level `GormConfig` used by adapters and by callers
that already have a `gorm.Dialector`.

## Install

This package is part of `github.com/phcp-tech/common-library-golang`.

```bash
go get github.com/phcp-tech/common-library-golang
```

If your application uses a specific adapter, make sure the matching GORM driver
is available in the application module:

```bash
go get gorm.io/driver/mysql
go get gorm.io/driver/postgres
go get gorm.io/driver/sqlite
```

## Opening a Database

### MySQL

```go
package main

import (
    "github.com/phcp-tech/common-library-golang/dbgorm/mysql"
    "gorm.io/gorm"
)

func openMySQL() (*gorm.DB, error) {
    return mysql.NewMySQL(mysql.Config{
        Host:            "localhost",
        Port:            "3306",
        Database:        "risk",
        Username:        "risk",
        Password:        "secret",
        MaxOpenConns:    50,
        MaxIdleConns:    10,
        ConnMaxLifetime: 60,
        ConnMaxIdletime: 10,
    })
}
```

### PostgreSQL

```go
db, err := postgres.NewPostgres(postgres.Config{
    Host:            "localhost",
    Port:            "5432",
    Database:        "risk",
    Username:        "risk",
    Password:        "secret",
    SearchPath:      "public",
    MaxOpenConns:    50,
    MaxIdleConns:    10,
    ConnMaxLifetime: 60,
    ConnMaxIdletime: 10,
})
```

### SQLite

```go
db, err := sqlite.NewSQLite(sqlite.Config{
    Path: "file:risk.db?cache=shared",
})
```

SQLite always opens through `dbgorm.Open` with `MaxOpenConns` and
`MaxIdleConns` set to `1` to reduce `database is locked` errors. Its adapter
config accepts only `Path` and optional `Logger`.

For tests:

```go
db, err := sqlite.NewSQLite(sqlite.Config{
    Path: "file::memory:?cache=shared",
})
```

### Low-level Open

Use `dbgorm.Open` only when you already have a `gorm.Dialector`:

```go
db, err := dbgorm.Open(dialector, &dbgorm.GormConfig{
    MaxOpenConns:    50,
    MaxIdleConns:    10,
    ConnMaxLifetime: 60,
    ConnMaxIdletime: 10,
    Logger:          slog.Default(),
})
```

`ConnMaxLifetime` and `ConnMaxIdletime` are expressed in minutes. If a
`GormConfig` field is zero, `dbgorm.Open` uses the library default.

## Default Instance

Use the default instance when the process has one primary database connection.

```go
if err := mysql.InitDefault(mysql.Config{
    Host:     "localhost",
    Port:     "3306",
    Database: "risk",
    Username: "risk",
    Password: "secret",
}); err != nil {
    return err
}

db := dbgorm.Default()
```

The default instance is process-wide. It is not one default per driver. If an
application uses multiple databases, prefer explicit `*gorm.DB` fields:

```go
type Store struct {
    MainDB    *gorm.DB
    MetricsDB *gorm.DB
}
```

Useful default-instance helpers:

```go
db := dbgorm.Default()
dbgorm.SetDefault(db)
```

Applications should close or ping the default database through the returned
`*gorm.DB`:

```go
db := dbgorm.Default()
if db != nil {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    if err := sqlDB.PingContext(ctx); err != nil {
        return err
    }
    defer sqlDB.Close()
}
```

## Logging

By default, `Open` configures GORM SQL logging with the standard library
`slog.Default()` logger.

To use the project logger, pass its `*slog.Logger` through `Config.Logger`:

```go
import applog "github.com/phcp-tech/common-library-golang/log"

db, err := mysql.NewMySQL(mysql.Config{
    Host:     "localhost",
    Port:     "3306",
    Database: "risk",
    Username: "risk",
    Password: "secret",
    Logger:   applog.Instance().Logger,
})
```

To inspect the logger formats used by the package tests, run:

```bash
go test -count=1 -v ./dbgorm -run "TestOpenUses(DefaultSlogLogger|LibraryLogInstance)"
```

`TestOpenUsesDefaultSlogLogger` prints the JSON output captured from
`slog.Default()`. `TestOpenUsesLibraryLogInstance` triggers the JSON slog
handler configured by `library-golang/log/slog.go`.

## Normal GORM Usage

After opening a database, use standard GORM APIs directly:

```go
err := db.WithContext(ctx).Create(&user).Error
```

```go
err := db.WithContext(ctx).
    Where("status = ?", "active").
    Order("created_at desc").
    Find(&users).Error
```

Transactions stay as normal GORM transactions:

```go
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&order).Error; err != nil {
        return err
    }

    return dbgorm.DeleteWhere(ctx, tx, &CartItem{}, "user_id = ?", userID)
})
```

## Helper Functions

### Find by ID

```go
user, err := dbgorm.FirstByID[User](ctx, db, userID)
if err != nil {
    if dbgorm.IsNotFound(err) {
        return nil, ErrUserNotFound
    }
    return nil, err
}
```

### Find by Condition

```go
user, err := dbgorm.FirstWhere[User](ctx, db, "email = ?", email)
```

### Delete by ID

`DeleteByID` returns `gorm.ErrRecordNotFound` when no row is deleted.

```go
err := dbgorm.DeleteByID[User](ctx, db, userID)
if dbgorm.IsNotFound(err) {
    return ErrUserNotFound
}
```

### Delete by Condition

`DeleteWhere` treats zero affected rows as success.

```go
err := dbgorm.DeleteWhere(ctx, db, &Session{}, "expires_at < ?", cutoff)
```

## Pagination and Sorting

Use scopes for reusable query modifiers:

```go
allowedSorts := map[string]string{
    "id":        "id",
    "createdAt": "created_at",
    "email":     "email",
}

err := db.WithContext(ctx).
    Scopes(
        dbgorm.OrderBy(allowedSorts, req.Sort, req.Direction),
        dbgorm.Paginate(req.Page, req.Limit),
    ).
    Find(&users).Error
```

`OrderBy` only accepts columns listed in the allow-list. This avoids directly
injecting request values into `ORDER BY`.

## Raw SQL

Use raw SQL helpers when a query is easier to express manually. Always pass
values through arguments instead of string concatenation.

```go
rows, err := dbgorm.ExecRaw(
    ctx,
    db,
    "update users set status = ? where id = ?",
    "disabled",
    userID,
)
```

```go
var total int64
err := dbgorm.ScanRaw(
    ctx,
    db,
    &total,
    "select count(*) from users where status = ?",
    "active",
)
```

## AutoMigrate

```go
err := dbgorm.AutoMigrate(ctx, db, dbgorm.MigrateOptions{
    Enabled: true,
}, &User{}, &Order{})
```

Mock SQL can be executed after migration:

```go
err := dbgorm.AutoMigrate(ctx, db, dbgorm.MigrateOptions{
    Enabled:    true,
    InsertMock: true,
    MockFile:   "testdata/mock.sql",
    Separator:  "----",
}, &User{})
```

`AutoMigrate` does not read environment variables. Applications are responsible
for deciding when migration and mock data are enabled.

## Error Handling

Use `IsNotFound` to check for GORM's not-found condition:

```go
if dbgorm.IsNotFound(err) {
    // map to 404, domain error, or another application-specific response
}
```

Common package errors:

```go
dbgorm.ErrNilDialector
dbgorm.ErrMissingConfig
```

## Configuration Ownership

`dbgorm` does not read application environment variables. Load config in the
application, then build the adapter config explicitly:

```go
db, err := mysql.NewMySQL(mysql.Config{
    Host:            env.DBHost,
    Port:            env.DBPort,
    Database:        env.DBName,
    Username:        env.DBUser,
    Password:        env.DBPassword,
    MaxOpenConns:    env.DBMaxOpenConns,
    MaxIdleConns:    env.DBMaxIdleConns,
    ConnMaxLifetime: env.DBConnMaxLifetime,
    ConnMaxIdletime: env.DBConnMaxIdletime,
    Logger:          appLogger,
})
```

## Testing

SQLite in-memory is convenient for tests:

```go
db, err := sqlite.NewSQLite(sqlite.Config{
    Path: "file::memory:?cache=shared",
})
```

Run package tests:

```bash
go test ./dbgorm/...
```
