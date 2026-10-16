-- +goose Up
-- Ringkasan AI: nomor versi per siswa + catatan (mis. "Menambahkan data terbaru").
ALTER TABLE ai_summaries ADD COLUMN version INT NOT NULL DEFAULT 1;
ALTER TABLE ai_summaries ADD COLUMN note TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE ai_summaries DROP COLUMN note;
ALTER TABLE ai_summaries DROP COLUMN version;
