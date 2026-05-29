-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET updated_at = $2, email = $3, hashed_password = $4
WHERE id = $1
RETURNING *;

-- name: MakeUserRed :exec
UPDATE users
SET is_chirpy_red = true
WHERE id = $1;