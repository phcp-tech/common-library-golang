-- Example queries. sqlc generates type-safe Go functions from these annotations.
-- Annotation format:  -- name: FunctionName :one|:many|:exec|:execresult|:execrows

-- name: GetUser :one
SELECT id, login, name, `group`, tag, created_at, updated_at
FROM users
WHERE login = ?
LIMIT 1;

-- name: ListUsers :many
SELECT id, login, name, `group`, tag, created_at, updated_at
FROM users
ORDER BY login;

-- name: CreateUser :execresult
INSERT INTO users (login, name, `group`, tag)
VALUES (?, ?, ?, ?);

-- name: UpdateUserTag :exec
UPDATE users
SET tag = ?
WHERE login = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE login = ?;
