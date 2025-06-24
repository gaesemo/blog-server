-- name: CreateUser :one
INSERT INTO users (
    identity_provider,
    email,
    username,
    avatar_url,
    about_me,
    created_at,
    updated_at
) VALUES (
    $1,$2,$3,$4,$5,$6,$7
)
RETURNING *;

-- name: GetUserById :one
SELECT * 
FROM users
WHERE deleted_at IS NULL AND id = $1;

-- name: GetUserByEmailAndIDP :one
SELECT *
FROM users
WHERE deleted_at IS NULL
AND email = $1
AND identity_provider = $2;