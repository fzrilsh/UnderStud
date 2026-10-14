-- name: ListObservationsByStudent :many
SELECT o.*, u.name AS author_name
FROM observations o
JOIN users u ON u.id = o.author_id
WHERE o.student_id = $1
ORDER BY o.tanggal DESC, o.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateObservation :one
INSERT INTO observations (
  student_id, author_id, items, categories, red_flag, catatan, tanggal, tindakan_awal
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetObservationByID :one
SELECT * FROM observations WHERE id = $1;

-- name: UpdateObservation :one
UPDATE observations SET
  items = $2, categories = $3, red_flag = $4, catatan = $5,
  tanggal = $6, tindakan_awal = $7, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateObservationRevision :one
INSERT INTO observation_revisions (observation_id, edited_by, snapshot)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListObservationRevisions :many
SELECT * FROM observation_revisions WHERE observation_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: HandleRedFlag :one
UPDATE observations SET handled_at = now(), handled_by = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListActiveRedFlagsScoped :many
SELECT o.id, o.student_id, o.items, o.tanggal, o.created_at, s.name AS student_name, s.kelas
FROM observations o
JOIN students s ON s.id = o.student_id
WHERE o.red_flag = true AND o.handled_at IS NULL
  AND (sqlc.arg(kelas)::text = '' OR s.kelas = sqlc.arg(kelas))
ORDER BY o.tanggal DESC, o.created_at DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountActiveRedFlagsScoped :one
SELECT count(*) FROM observations o
JOIN students s ON s.id = o.student_id
WHERE o.red_flag = true AND o.handled_at IS NULL
  AND (sqlc.arg(kelas)::text = '' OR s.kelas = sqlc.arg(kelas));

-- name: StudentHasRedFlag :one
-- NEW-1: cek red flag belum ditangani via EXISTS (tanpa LIMIT 200, tanpa N+1).
SELECT EXISTS (
  SELECT 1 FROM observations
  WHERE student_id = $1 AND red_flag = true AND handled_at IS NULL
) AS has_red_flag;

-- name: ListRedFlagStudentIDs :many
-- NEW-1: satu query untuk semua siswa ter-scope (hindari N+1 di list).
SELECT DISTINCT student_id FROM observations
WHERE red_flag = true AND handled_at IS NULL AND student_id = ANY(sqlc.arg(ids)::uuid[]);
