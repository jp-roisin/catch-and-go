-- name: ListStopsFromLine :many
SELECT s.*
FROM stops_by_lines sbl
JOIN stops s ON s.id = sbl.stop_id
WHERE sbl.line_id = ?
ORDER BY sbl."order" ASC;
