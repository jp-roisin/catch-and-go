-- name: GetLine :one
SELECT * FROM lines
WHERE code = ? AND direction = ? LIMIT 1;

-- name: ListLines :many
SELECT * FROM lines
ORDER BY code ASC;

-- name: ListLinesByDirection :many
SELECT * FROM lines
WHERE direction = ?
ORDER BY code ASC;

-- name: ListLinesByCode :many
SELECT * FROM lines
WHERE CODE = ?;
