-- name: GetUserById :one
SELECT * 
FROM users
WHERE deleted_at IS NULL AND id = $1;