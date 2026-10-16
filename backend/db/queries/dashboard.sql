-- name: CountChatsByStatus :one
-- Metrik kartu "Chat bulan ini" (dashboard BK): semua hitungan sisi staf,
-- chat yang di-hide staf dikecualikan.
SELECT
  COUNT(*) FILTER (WHERE status = 'baru' AND hidden_by_staff_at IS NULL)                          AS waiting,
  COUNT(*) FILTER (WHERE pinned_at IS NOT NULL AND drafted_at IS NULL AND status <> 'dilewati'
                     AND hidden_by_staff_at IS NULL)                                              AS pinned,
  COUNT(*) FILTER (WHERE drafted_at IS NOT NULL AND hidden_by_staff_at IS NULL)                   AS drafted,
  COUNT(*) FILTER (WHERE status IN ('baru','dibuka') AND pinned_at IS NULL AND drafted_at IS NULL
                     AND hidden_by_staff_at IS NULL)                                              AS queued,
  COUNT(*) FILTER (WHERE status = 'baru' AND pinned_at IS NULL AND drafted_at IS NULL
                     AND hidden_by_staff_at IS NULL)                                              AS unqueued
FROM chats;

-- name: CountActiveStudents :one
-- Ringkasan siswa non-arsip: berapa yang red flag & berapa yang punya ringkasan.
SELECT
  COUNT(*) FILTER (WHERE archived_at IS NULL AND EXISTS (
      SELECT 1 FROM observations o
      WHERE o.student_id = students.id AND o.red_flag = true AND o.handled_at IS NULL
  ))                                                                                AS red_flags,
  COUNT(*) FILTER (WHERE archived_at IS NULL AND EXISTS (
      SELECT 1 FROM ai_summaries s WHERE s.student_id = students.id AND s.status = 'ready'
  ))                                                                                AS summaries
FROM students;

-- name: ListClassStats :many
-- Statistik per kelas (semua kelas, termasuk yang kosong) untuk tabel "Data kelas".
SELECT
  st.kelas,
  COUNT(*)                                                              AS students,
  COUNT(*) FILTER (WHERE EXISTS (
      SELECT 1 FROM observations o
      WHERE o.student_id = st.id AND o.red_flag = true AND o.handled_at IS NULL
  ))                                                                    AS red_flag
FROM students st
WHERE st.archived_at IS NULL
GROUP BY st.kelas
ORDER BY st.kelas;

-- name: ListHomerooms :many
-- Nama wali kelas per kelas (untuk kolom "homeroom" tabel Data kelas).
SELECT kelas, name FROM users
WHERE role = 'WALI_KELAS' AND kelas IS NOT NULL AND is_active = true;

-- name: CountChatsByMonth :many
-- Seri chart 6 bulan: chat masuk per bulan + status saat ini.
SELECT
  (date_trunc('month', created_at))::timestamptz                          AS bulan,
  COUNT(*) FILTER (WHERE pinned_at IS NOT NULL)                           AS pinned,
  COUNT(*) FILTER (WHERE drafted_at IS NOT NULL)                          AS drafted,
  COUNT(*) FILTER (WHERE status IN ('baru','dibuka'))                     AS queued
FROM chats
WHERE created_at >= date_trunc('month', now()) - interval '5 months'
GROUP BY bulan
ORDER BY bulan;

-- name: ListLatestSummaries :many
-- 4 ringkasan terbaru (join siswa) untuk kartu "Ringkasan terbaru".
SELECT s.*, st.name AS student_name, st.kelas AS student_kelas, st.archived_at AS student_archived_at,
  EXISTS (SELECT 1 FROM observations o
          WHERE o.student_id = st.id AND o.red_flag = true AND o.handled_at IS NULL) AS student_has_red_flag
FROM ai_summaries s
JOIN students st ON st.id = s.student_id
WHERE st.archived_at IS NULL AND s.status = 'ready'
ORDER BY s.created_at DESC
LIMIT $1;

-- name: GetRedTopStudent :one
-- Siswa red flag non-arsip pertama (banner "perlu segera ditinjau").
SELECT st.*, o.tanggal AS red_tanggal
FROM students st
JOIN observations o ON o.student_id = st.id
WHERE st.archived_at IS NULL AND o.red_flag = true AND o.handled_at IS NULL
ORDER BY o.tanggal DESC, o.created_at DESC
LIMIT 1;

-- name: CountStudentsByKelas :one
-- Statistik kelas untuk dashboard Wali Kelas.
SELECT
  COUNT(*)                                                                        AS students,
  COUNT(*) FILTER (WHERE EXISTS (
      SELECT 1 FROM student_profiles p
      WHERE p.student_id = students.id
        AND jsonb_array_length(COALESCE(p.nilai_erapor, '[]'::jsonb)) >= 5
  ))                                                                              AS e_rapor_complete,
  COUNT(*) FILTER (WHERE EXISTS (
      SELECT 1 FROM observations o WHERE o.student_id = students.id
  ))                                                                              AS observed,
  COUNT(*) FILTER (WHERE EXISTS (
      SELECT 1 FROM observations o
      WHERE o.student_id = students.id AND o.red_flag = true AND o.handled_at IS NULL
  ))                                                                              AS red_flags
FROM students
WHERE kelas = $1 AND archived_at IS NULL;

-- name: GetRedTopStudentByKelas :one
SELECT st.*, o.tanggal AS red_tanggal
FROM students st
JOIN observations o ON o.student_id = st.id
WHERE st.archived_at IS NULL AND st.kelas = $1 AND o.red_flag = true AND o.handled_at IS NULL
ORDER BY o.tanggal DESC, o.created_at DESC
LIMIT 1;
