-- name: Createdashboard :one
INSERT INTO dashboards (
    session_id,
    stop_id
) VALUES (
    ?, ?
)
RETURNING *;
