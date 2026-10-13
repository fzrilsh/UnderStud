-- name: CreateAnonSession :one
INSERT INTO anon_sessions (token_hash) VALUES ($1) RETURNING *;

-- name: GetAnonSessionByTokenHash :one
SELECT * FROM anon_sessions WHERE token_hash = $1;

-- name: UpdateAnonProfile :one
UPDATE anon_sessions
SET display_name = $2, display_kelas = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListChatsBySession :many
SELECT * FROM chats
WHERE session_id = $1 AND hidden_by_student_at IS NULL
ORDER BY last_message_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateChat :one
INSERT INTO chats (session_id, identity_mode, snapshot_name, snapshot_kelas)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateMessage :one
INSERT INTO messages (chat_id, sender, sender_user_id, body)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CountChatsSince :one
SELECT count(*) FROM chats WHERE session_id = $1 AND created_at >= $2;

-- name: CountMessagesSince :one
SELECT count(*) FROM messages m
JOIN chats c ON c.id = m.chat_id
WHERE c.session_id = $1 AND m.sender = 'SISWA' AND m.created_at >= $2;
