-- +goose Up
-- NIS dibuang; NISN jadi identifier (leger E-Rapor pakai NISN). Spec revisi 15.
ALTER TABLE students RENAME COLUMN nis TO nisn;
ALTER INDEX students_nis_key RENAME TO students_nisn_key;

-- Rekap kehadiran dari kolom S/I/A leger.
ALTER TABLE student_profiles ADD COLUMN sakit INT NOT NULL DEFAULT 0;
ALTER TABLE student_profiles ADD COLUMN izin  INT NOT NULL DEFAULT 0;
ALTER TABLE student_profiles ADD COLUMN alpa  INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE student_profiles DROP COLUMN alpa;
ALTER TABLE student_profiles DROP COLUMN izin;
ALTER TABLE student_profiles DROP COLUMN sakit;
ALTER INDEX students_nisn_key RENAME TO students_nis_key;
ALTER TABLE students RENAME COLUMN nisn TO nis;
