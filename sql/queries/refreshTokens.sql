-- name: CreateToken :one
INSERT INTO refresh_tokens (token, user_id,  expires_at)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: UpdateToken :one
UPDATE refresh_tokens
SET token = $1, expires_at = $2, updated_at = now()
WHERE token = $3
RETURNING *;

-- name: RevokeToken :one
UPDATE refresh_tokens
SET revoked_at = $1, updated_at = now()
WHERE token = $2
RETURNING *;