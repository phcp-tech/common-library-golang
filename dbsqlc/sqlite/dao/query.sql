-- Example queries. sqlc generates type-safe Go functions from these annotations.
-- Annotation format:  -- name: FunctionName :one|:many|:exec|:execresult|:execrows

-- name: GetUser :one
SELECT id, login, name, group_name, tag, created_at, updated_at
FROM users
WHERE login = ?
LIMIT 1;

-- name: ListUsers :many
SELECT id, login, name, group_name, tag, created_at, updated_at
FROM users
ORDER BY login;

-- name: CreateUser :execresult
INSERT INTO users (login, name, group_name, tag)
VALUES (?, ?, ?, ?);

-- name: UpdateUserTag :exec
UPDATE users
SET tag = ?, updated_at = strftime('%Y-%m-%d %H:%M:%S', 'now')
WHERE login = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE login = ?;
