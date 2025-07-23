-- name: Createdashboard :one
INSERT INTO dashboards (
    session_id,
    stop_id
) VALUES (
    ?, ?
)
RETURNING *;


-- name: ListDashboardsFromSession :many
SELECT
  d.id AS dashboard_id,
  d.session_id,
  d.stop_id,
  d.created_at AS dashboard_created_at,
  s.id AS stop_id,
  s.code AS stop_code,
  s.geo AS stop_geo,
  s.name AS stop_name,
  s.created_at AS stop_created_at
FROM dashboards d
JOIN stops s ON s.id = d.stop_id
WHERE d.session_id = ?
ORDER BY s.created_at ASC;

-- name: DeleteDashboard :exec
DELETE from dashboards
WHERE id = ? AND session_id = ?;
