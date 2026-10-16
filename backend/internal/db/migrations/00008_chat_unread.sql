-- +goose Up
-- Jumlah balasan BK yang belum dibaca siswa (badge di sidebar siswa).
-- Dinaikkan setiap BK membalas, direset saat siswa membuka chat.
ALTER TABLE chats ADD COLUMN student_unread_count INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE chats DROP COLUMN student_unread_count;
