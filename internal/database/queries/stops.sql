-- name: GetStop :one
SELECT * FROM stops
WHERE code = ? LIMIT 1;

-- name: ListStops :many
SELECT * FROM stops
ORDER BY code ASC;
