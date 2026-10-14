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

-- name: ClaimImportPreview :one
-- F7: klaim token secara ATOMIK (hapus + kembalikan dalam satu statement).
-- Commit ganda/paralel dengan token sama → hanya satu yang berhasil.
DELETE FROM import_previews
WHERE token = $1 AND expires_at > now()
RETURNING *;

-- name: DeleteImportPreview :exec
DELETE FROM import_previews WHERE token = $1;

-- name: DeleteExpiredImportPreviews :exec
-- F24: bersihkan preview kedaluwarsa (yang di-abandon) agar tabel tak tumbuh.
DELETE FROM import_previews WHERE expires_at <= now();
