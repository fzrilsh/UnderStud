-- +goose Up
CREATE TABLE ai_summaries (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id     UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    generated_by   UUID NOT NULL REFERENCES users(id),
    trigger_reason TEXT NOT NULL,
    content        TEXT NOT NULL,
    sources        JSONB,
    model          TEXT,
    status         TEXT NOT NULL DEFAULT 'ready',
    error          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ai_summaries_status_check CHECK (status IN ('ready', 'failed'))
);
CREATE INDEX ai_summaries_student_idx ON ai_summaries (student_id, created_at DESC);

CREATE TABLE export_records (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    summary_id  UUID NOT NULL REFERENCES ai_summaries(id) ON DELETE CASCADE,
    exported_by UUID NOT NULL REFERENCES users(id),
    format      TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX export_records_summary_idx ON export_records (summary_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS export_records;
DROP TABLE IF EXISTS ai_summaries;
