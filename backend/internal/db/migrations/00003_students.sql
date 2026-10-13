-- +goose Up
CREATE TABLE students (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nis         TEXT UNIQUE,
    name        TEXT NOT NULL,
    kelas       TEXT NOT NULL,
    archived_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX students_kelas_archived_idx ON students (kelas, archived_at);
CREATE INDEX students_name_idx ON students (lower(name));

CREATE TABLE student_profiles (
    student_id             UUID PRIMARY KEY REFERENCES students(id) ON DELETE CASCADE,
    nilai_erapor           JSONB,
    minat_bakat            TEXT,
    tanggal_lahir          DATE,
    catatan_psikologis     TEXT,
    catatan_kasus_prestasi TEXT,
    masukan_bk             TEXT,
    pendapat_wali_kelas    TEXT,
    updated_by             UUID REFERENCES users(id),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS student_profiles;
DROP TABLE IF EXISTS students;
