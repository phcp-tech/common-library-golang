-- Example schema. Replace with your actual table definitions.
-- sqlc reads this file to understand column types for code generation.
-- SQLite uses dynamic typing; INTEGER PRIMARY KEY is implicitly an alias for rowid (autoincrement).

CREATE TABLE IF NOT EXISTS users (
    id         INTEGER  PRIMARY KEY,
    login      INTEGER  NOT NULL UNIQUE,
    name       TEXT     NOT NULL DEFAULT '',
    group_name TEXT     NOT NULL DEFAULT '',
    tag        TEXT,
    created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    updated_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now'))
);
