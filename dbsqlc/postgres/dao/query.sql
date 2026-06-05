-- Example queries. sqlc generates type-safe Go functions from these annotations.
-- PostgreSQL uses $1, $2, ... positional placeholders (not ? like MySQL/SQLite).
-- Annotation format:  -- name: FunctionName :one|:many|:exec|:execresult|:execrows

-- name: GetUser :one
SELECT id, login, name, group_name, tag, created_at, updated_at
FROM users
WHERE login = $1
LIMIT 1;

-- name: ListUsers :many
SELECT id, login, name, group_name, tag, created_at, updated_at
FROM users
ORDER BY login;

-- name: CreateUser :one
INSERT INTO users (login, name, group_name, tag)
VALUES ($1, $2, $3, $4)
RETURNING id, login, name, group_name, tag, created_at, updated_at;

-- name: UpdateUserTag :exec
UPDATE users
SET tag = $1, updated_at = NOW()
WHERE login = $2;

-- name: DeleteUser :exec
DELETE FROM users
WHERE login = $1;
