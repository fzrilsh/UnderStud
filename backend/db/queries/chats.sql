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

-- name: CountChatsBySession :one
SELECT count(*) FROM chats
WHERE session_id = $1 AND hidden_by_student_at IS NULL;

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

-- name: GetChatByID :one
SELECT * FROM chats WHERE id = $1;

-- name: GetChatWithSession :one
SELECT c.*, a.token_hash AS session_token_hash
FROM chats c
JOIN anon_sessions a ON a.id = c.session_id
WHERE c.id = $1;

-- name: ListMessagesByChat :many
SELECT * FROM messages WHERE chat_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3;

-- name: GetMessageByID :one
SELECT * FROM messages WHERE id = $1;

-- name: UpdateChatStatus :one
UPDATE chats SET status = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: PinChat :one
UPDATE chats SET pinned_at = now(), drafted_at = NULL, updated_at = now() WHERE id = $1 RETURNING *;

-- name: UnpinChat :one
UPDATE chats SET pinned_at = NULL, updated_at = now() WHERE id = $1 RETURNING *;

-- name: DraftChat :one
UPDATE chats SET drafted_at = now(), pinned_at = NULL, updated_at = now() WHERE id = $1 RETURNING *;

-- name: UndraftChat :one
UPDATE chats SET drafted_at = NULL, updated_at = now() WHERE id = $1 RETURNING *;

-- name: SetChatSkipped :one
-- Next (dari section mana pun): keluar sidebar total. Audit H1: WAJIB lepas
-- pinned_at juga (Next dari Pin), selain drafted_at + status='dilewati'.
UPDATE chats SET pinned_at = NULL, drafted_at = NULL, status = 'dilewati', updated_at = now()
WHERE id = $1 RETURNING *;

-- name: MarkChatOpened :one
-- Transisi baru→dibuka saat BK pertama membuka laporan (spec §7.1). Audit H2.
UPDATE chats SET status = 'dibuka', updated_at = now()
WHERE id = $1 AND status = 'baru' RETURNING *;

-- name: HideByStudent :one
UPDATE chats SET hidden_by_student_at = now(), pinned_at = NULL, drafted_at = NULL, updated_at = now() WHERE id = $1 RETURNING *;

-- name: HideByStaff :one
UPDATE chats SET hidden_by_staff_at = now(), drafted_at = NULL, updated_at = now() WHERE id = $1 RETURNING *;

-- name: TouchChatMessage :one
UPDATE chats SET last_message_at = now(), updated_at = now() WHERE id = $1 RETURNING *;

-- name: ListStaffQueue :many
-- Sidebar BK = PIN ∪ DRAFT ∪ ANTREAN (spec §7.2). Chat 'dilewati' TIDAK masuk
-- section mana pun → keluar dari sidebar. Pin/Draft selalu berstatus baru/dibuka,
-- jadi filter status cukup untuk ketiganya.
SELECT * FROM chats
WHERE hidden_by_staff_at IS NULL
  AND hidden_by_student_at IS NULL
  AND status IN ('baru', 'dibuka')
ORDER BY (pinned_at IS NOT NULL) DESC, (drafted_at IS NOT NULL) DESC,
         pinned_at DESC NULLS LAST, drafted_at DESC NULLS LAST,
         last_message_at DESC
LIMIT $1 OFFSET $2;

-- name: CountStaffQueue :one
SELECT count(*) FROM chats
WHERE hidden_by_staff_at IS NULL
  AND hidden_by_student_at IS NULL
  AND status IN ('baru', 'dibuka');

-- name: ListStaffHistory :many
SELECT * FROM chats
WHERE status = 'dilewati' OR hidden_by_staff_at IS NOT NULL
ORDER BY last_message_at DESC
LIMIT $1 OFFSET $2;

-- name: CountStaffHistory :one
SELECT count(*) FROM chats
WHERE status = 'dilewati' OR hidden_by_staff_at IS NOT NULL;

-- name: HasStaffReplied :one
SELECT EXISTS (
  SELECT 1 FROM messages WHERE chat_id = $1 AND sender = 'BK'
) AS has_replied;

-- name: ListLastMessageSenders :many
-- F23: pengirim pesan TERAKHIR per chat (batch) — turunkan has_new_reply
-- (siswa) & unread_by_staff (BK) tanpa tabel read-state baru.
SELECT DISTINCT ON (m.chat_id) m.chat_id, m.sender
FROM messages m
WHERE m.chat_id = ANY(sqlc.arg(chat_ids)::uuid[])
ORDER BY m.chat_id, m.created_at DESC, m.id DESC;

-- name: CreateAttachment :one
INSERT INTO attachments (message_id, stored_path, original_name, mime, size_bytes)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAttachmentByID :one
SELECT * FROM attachments WHERE id = $1;

-- name: ListAttachmentsByMessage :many
SELECT * FROM attachments WHERE message_id = $1 ORDER BY created_at ASC;

-- name: ListAttachmentsByMessages :many
-- F4: ambil lampiran untuk BANYAK pesan dalam satu query (hindari N+1).
SELECT * FROM attachments
WHERE message_id = ANY(sqlc.arg(message_ids)::uuid[])
ORDER BY created_at ASC;
