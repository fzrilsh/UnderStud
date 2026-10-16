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
-- `filter` (opsional): all | red_flag | archived | incomplete | not_observed.
-- Status per-siswa (e_rapor_complete/observed/has_red_flag) dihitung via EXISTS
-- agar UI bisa menampilkan kolom status tanpa request tambahan.
SELECT
  st.*,
  EXISTS (SELECT 1 FROM observations o
          WHERE o.student_id = st.id AND o.red_flag = true AND o.handled_at IS NULL) AS has_red_flag,
  EXISTS (SELECT 1 FROM student_profiles p
          WHERE p.student_id = st.id
            AND jsonb_array_length(COALESCE(p.nilai_erapor, '[]'::jsonb)) >= 5)      AS e_rapor_complete,
  EXISTS (SELECT 1 FROM observations o WHERE o.student_id = st.id)                   AS observed,
  EXISTS (SELECT 1 FROM ai_summaries s
          WHERE s.student_id = st.id AND s.status = 'ready')                         AS has_summary
FROM students st
WHERE (sqlc.arg(kelas)::text = '' OR st.kelas = sqlc.arg(kelas))
  -- filter='archived' meng-override include_archived (hanya tampilkan yang diarsipkan).
  AND (
    sqlc.arg(filter)::text = 'archived'
    OR sqlc.arg(include_archived)::bool = true
    OR st.archived_at IS NULL
  )
  AND (sqlc.arg(search)::text = '' OR lower(st.name) LIKE '%' || lower(sqlc.arg(search)) || '%' OR st.nisn LIKE '%' || sqlc.arg(search) || '%')
  AND (
    sqlc.arg(filter)::text = '' OR sqlc.arg(filter)::text = 'all'
    OR (sqlc.arg(filter)::text = 'red_flag' AND EXISTS (
        SELECT 1 FROM observations o
        WHERE o.student_id = st.id AND o.red_flag = true AND o.handled_at IS NULL))
    OR (sqlc.arg(filter)::text = 'archived' AND st.archived_at IS NOT NULL)
    OR (sqlc.arg(filter)::text = 'incomplete' AND NOT EXISTS (
        SELECT 1 FROM student_profiles p
        WHERE p.student_id = st.id
          AND jsonb_array_length(COALESCE(p.nilai_erapor, '[]'::jsonb)) >= 5))
    OR (sqlc.arg(filter)::text = 'not_observed' AND NOT EXISTS (
        SELECT 1 FROM observations o WHERE o.student_id = st.id))
  )
ORDER BY st.name ASC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountStudents :one
SELECT count(*) FROM students st
WHERE (sqlc.arg(kelas)::text = '' OR st.kelas = sqlc.arg(kelas))
  -- filter='archived' meng-override include_archived (konsisten dengan ListStudents).
  AND (
    sqlc.arg(filter)::text = 'archived'
    OR sqlc.arg(include_archived)::bool = true
    OR st.archived_at IS NULL
  )
  AND (sqlc.arg(search)::text = '' OR lower(st.name) LIKE '%' || lower(sqlc.arg(search)) || '%' OR st.nisn LIKE '%' || sqlc.arg(search) || '%')
  AND (
    sqlc.arg(filter)::text = '' OR sqlc.arg(filter)::text = 'all'
    OR (sqlc.arg(filter)::text = 'red_flag' AND EXISTS (
        SELECT 1 FROM observations o
        WHERE o.student_id = st.id AND o.red_flag = true AND o.handled_at IS NULL))
    OR (sqlc.arg(filter)::text = 'archived' AND st.archived_at IS NOT NULL)
    OR (sqlc.arg(filter)::text = 'incomplete' AND NOT EXISTS (
        SELECT 1 FROM student_profiles p
        WHERE p.student_id = st.id
          AND jsonb_array_length(COALESCE(p.nilai_erapor, '[]'::jsonb)) >= 5))
    OR (sqlc.arg(filter)::text = 'not_observed' AND NOT EXISTS (
        SELECT 1 FROM observations o WHERE o.student_id = st.id))
  );

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
