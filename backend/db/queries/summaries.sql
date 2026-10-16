-- name: CreateSummary :one
INSERT INTO ai_summaries (student_id, generated_by, trigger_reason, content, sources, model, status, error, version, note)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: CountSummaries :one
SELECT count(*) FROM ai_summaries WHERE student_id = $1;

-- name: NextSummaryVersion :one
-- Nomor versi berikutnya untuk siswa (1 bila belum ada ringkasan).
SELECT COALESCE(MAX(version), 0) + 1 AS next_version FROM ai_summaries WHERE student_id = $1;

-- name: GetLatestSummary :one
SELECT s.*, u.name AS author_name
FROM ai_summaries s
JOIN users u ON u.id = s.generated_by
WHERE s.student_id = $1
ORDER BY s.created_at DESC
LIMIT 1;

-- name: ListSummaries :many
SELECT s.*, u.name AS author_name
FROM ai_summaries s
JOIN users u ON u.id = s.generated_by
WHERE s.student_id = $1
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetSummaryByID :one
SELECT s.*, u.name AS author_name
FROM ai_summaries s
JOIN users u ON u.id = s.generated_by
WHERE s.id = $1;
