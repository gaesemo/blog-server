-- name: GetProfile :one
SELECT id, username, email, avatar_url 
FROM users
WHERE deleted_at IS NULL AND id = $1;