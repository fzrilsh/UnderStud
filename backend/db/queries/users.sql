-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateStaff :one
INSERT INTO users (email, password_hash, name, role, kelas)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeactivateStaff :exec
UPDATE users SET is_active = false WHERE email = $1;
