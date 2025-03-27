-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, user_id, body)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpByID :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirpByID :exec
DELETE FROM chirps
WHERE id = $1
RETURNING *;

-- name: GetChirpsByUserID :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY created_at ASC;