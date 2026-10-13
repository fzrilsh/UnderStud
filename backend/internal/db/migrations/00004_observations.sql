-- +goose Up
CREATE TABLE observations (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id    UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    author_id     UUID NOT NULL REFERENCES users(id),
    items         JSONB NOT NULL,
    categories    JSONB NOT NULL,
    red_flag      BOOLEAN NOT NULL DEFAULT false,
    handled_at    TIMESTAMPTZ,
    handled_by    UUID REFERENCES users(id),
    catatan       TEXT,
    tanggal       DATE NOT NULL,
    tindakan_awal TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX observations_student_idx ON observations (student_id, tanggal DESC);
CREATE INDEX observations_redflag_idx ON observations (red_flag, handled_at) WHERE red_flag;

CREATE TABLE observation_revisions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    observation_id UUID NOT NULL REFERENCES observations(id) ON DELETE CASCADE,
    edited_by      UUID NOT NULL REFERENCES users(id),
    snapshot       JSONB NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX observation_revisions_obs_idx ON observation_revisions (observation_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS observation_revisions;
DROP TABLE IF EXISTS observations;
