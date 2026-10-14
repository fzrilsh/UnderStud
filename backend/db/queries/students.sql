-- name: GetStudentByID :one
SELECT * FROM students WHERE id = $1;

-- name: GetStudentByNISN :one
SELECT * FROM students WHERE nisn = $1;

-- name: CreateStudent :one
INSERT INTO students (nisn, name, kelas)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateStudentProfileEmpty :exec
INSERT INTO student_profiles (student_id) VALUES ($1)
ON CONFLICT (student_id) DO NOTHING;

-- name: CountObservationsByStudent :one
SELECT count(*) FROM observations WHERE student_id = $1;

-- name: ListStudents :many
SELECT * FROM students
WHERE (sqlc.arg(kelas)::text = '' OR kelas = sqlc.arg(kelas))
  AND (sqlc.arg(include_archived)::bool = true OR archived_at IS NULL)
  AND (sqlc.arg(search)::text = '' OR lower(name) LIKE '%' || lower(sqlc.arg(search)) || '%' OR nisn LIKE '%' || sqlc.arg(search) || '%')
ORDER BY name ASC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountStudents :one
SELECT count(*) FROM students
WHERE (sqlc.arg(kelas)::text = '' OR kelas = sqlc.arg(kelas))
  AND (sqlc.arg(include_archived)::bool = true OR archived_at IS NULL)
  AND (sqlc.arg(search)::text = '' OR lower(name) LIKE '%' || lower(sqlc.arg(search)) || '%' OR nisn LIKE '%' || sqlc.arg(search) || '%');

-- name: UpdateStudent :one
UPDATE students
SET nisn = $2, name = $3, kelas = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ArchiveStudent :one
UPDATE students SET archived_at = now(), updated_at = now() WHERE id = $1 RETURNING *;

-- name: RestoreStudent :one
UPDATE students SET archived_at = NULL, updated_at = now() WHERE id = $1 RETURNING *;

-- name: GetStudentProfile :one
SELECT * FROM student_profiles WHERE student_id = $1;

-- name: GetStudentProfileForUpdate :one
-- F7: kunci baris profil untuk read-modify-write (gabung nilai_erapor) agar
-- dua commit/update paralel tak saling menimpa (lost update).
SELECT * FROM student_profiles WHERE student_id = $1 FOR UPDATE;

-- name: UpsertStudentProfile :one
INSERT INTO student_profiles (
  student_id, nilai_erapor, sakit, izin, alpa, minat_bakat, tanggal_lahir,
  catatan_psikologis, catatan_kasus_prestasi, masukan_bk, pendapat_wali_kelas,
  updated_by, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, now())
ON CONFLICT (student_id) DO UPDATE SET
  nilai_erapor           = EXCLUDED.nilai_erapor,
  sakit                  = EXCLUDED.sakit,
  izin                   = EXCLUDED.izin,
  alpa                   = EXCLUDED.alpa,
  minat_bakat            = EXCLUDED.minat_bakat,
  tanggal_lahir          = EXCLUDED.tanggal_lahir,
  catatan_psikologis     = EXCLUDED.catatan_psikologis,
  catatan_kasus_prestasi = EXCLUDED.catatan_kasus_prestasi,
  masukan_bk             = EXCLUDED.masukan_bk,
  pendapat_wali_kelas    = EXCLUDED.pendapat_wali_kelas,
  updated_by             = EXCLUDED.updated_by,
  updated_at             = now()
RETURNING *;
