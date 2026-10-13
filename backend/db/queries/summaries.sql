-- name: CreateSummary :one
INSERT INTO ai_summaries (student_id, generated_by, trigger_reason, content, sources, model, status, error)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetLatestSummary :one
SELECT * FROM ai_summaries
WHERE student_id = $1
ORDER BY created_at DESC
LIMIT 1;
