-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetUserPassword :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUserDetails :one
UPDATE users SET (email, hashed_password, updated_at) = ($1, $2, NOW())
WHERE id = $3
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users;    

