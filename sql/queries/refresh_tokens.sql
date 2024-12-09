-- name: CreateRefToken :one
INSERT INTO refresh_tokens (
    token ,
    created_at ,
    updated_at ,   
    expires_at ,  
    user_id )
VALUES (
    $1,
    NOW(),
    NOW(),
    $3,
    $2
)
RETURNING *;


-- name: GetUserFromRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;


 
-- name: RevokeRefreshToken :one
UPDATE refresh_tokens SET (updated_at, revoked_at) = (NOW(), NOW())
WHERE token = $1
RETURNING expires_at;


 