-- name: GetSession :one
SELECT * FROM sessions
WHERE id = ? LIMIT 1;

-- name: ListSessions :many
SELECT * FROM sessions
ORDER BY created_at DESC;

-- name: CreateSession :one
INSERT INTO sessions (
    id
) VALUES (
    ?
)
RETURNING *;

-- name: DeleteSession :exec
DELETE from sessions
WHERE id = ?;

-- name: UpdateLocale :exec
UPDATE sessions
set locale = ?
WHERE id = ?;
