-- +goose Up
CREATE TABLE anon_sessions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash    TEXT NOT NULL UNIQUE,
    display_name  TEXT,
    display_kelas TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE chats (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id           UUID NOT NULL REFERENCES anon_sessions(id) ON DELETE CASCADE,
    status               TEXT NOT NULL DEFAULT 'baru',
    identity_mode        TEXT NOT NULL DEFAULT 'anonim',
    snapshot_name        TEXT,
    snapshot_kelas       TEXT,
    pinned_at            TIMESTAMPTZ,
    drafted_at           TIMESTAMPTZ,
    hidden_by_student_at TIMESTAMPTZ,
    hidden_by_staff_at   TIMESTAMPTZ,
    last_message_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chats_status_check CHECK (status IN ('baru', 'dibuka', 'dilewati')),
    CONSTRAINT chats_identity_check CHECK (identity_mode IN ('anonim', 'profil')),
    CONSTRAINT chats_pin_draft_exclusive CHECK (pinned_at IS NULL OR drafted_at IS NULL)
);
CREATE INDEX chats_queue_idx ON chats (status, pinned_at, drafted_at, last_message_at DESC);
CREATE INDEX chats_session_idx ON chats (session_id, hidden_by_student_at);

CREATE TABLE messages (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id        UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender         TEXT NOT NULL,
    sender_user_id UUID REFERENCES users(id),
    body           TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT messages_sender_check CHECK (sender IN ('SISWA', 'BK'))
);
CREATE INDEX messages_chat_idx ON messages (chat_id, created_at);

CREATE TABLE attachments (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id    UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    stored_path   TEXT NOT NULL,
    original_name TEXT,
    mime          TEXT NOT NULL,
    size_bytes    INT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX attachments_message_idx ON attachments (message_id);

CREATE TABLE import_previews (
    token      TEXT PRIMARY KEY,
    created_by UUID NOT NULL REFERENCES users(id),
    payload    JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS import_previews;
DROP TABLE IF EXISTS attachments;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chats;
DROP TABLE IF EXISTS anon_sessions;
