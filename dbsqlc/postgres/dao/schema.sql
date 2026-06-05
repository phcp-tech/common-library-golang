-- Example schema. Replace with your actual table definitions.
-- sqlc reads this file to understand column types for code generation.
-- PostgreSQL uses SERIAL/BIGSERIAL for autoincrement, and $1/$2/... for placeholders.

CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL    PRIMARY KEY,
    login      BIGINT       NOT NULL UNIQUE,
    name       TEXT         NOT NULL DEFAULT '',
    group_name TEXT         NOT NULL DEFAULT '',
    tag        TEXT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
