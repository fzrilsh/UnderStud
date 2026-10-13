-- name: ListActiveRedFlags :many
SELECT o.id, o.student_id, o.items, o.tanggal, s.name AS student_name, s.kelas
FROM observations o
JOIN students s ON s.id = o.student_id
WHERE o.red_flag = true AND o.handled_at IS NULL
ORDER BY o.tanggal DESC
LIMIT $1 OFFSET $2;

-- name: CountActiveRedFlags :one
SELECT count(*) FROM observations WHERE red_flag = true AND handled_at IS NULL;
