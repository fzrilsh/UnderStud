-- name: GetStudentByID :one
SELECT * FROM students WHERE id = $1;

-- name: GetStudentByNIS :one
SELECT * FROM students WHERE nis = $1;

-- name: CreateStudent :one
INSERT INTO students (nis, name, kelas)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateStudentProfileEmpty :exec
INSERT INTO student_profiles (student_id) VALUES ($1)
ON CONFLICT (student_id) DO NOTHING;

-- name: CountObservationsByStudent :one
SELECT count(*) FROM observations WHERE student_id = $1;
