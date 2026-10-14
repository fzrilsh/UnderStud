-- name: CreateExportRecord :one
INSERT INTO export_records (summary_id, exported_by, format)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateImportPreview :one
INSERT INTO import_previews (token, created_by, payload, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetImportPreview :one
SELECT * FROM import_previews WHERE token = $1 AND expires_at > now();

-- name: DeleteImportPreview :exec
DELETE FROM import_previews WHERE token = $1;
