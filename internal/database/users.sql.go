// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, email, hashed_password, is_chirpy_red
`

type CreateUserParams struct {
	Email          string
	HashedPassword string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.Email, arg.HashedPassword)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
		&i.IsChirpyRed,
	)
	return i, err
}

const deleteAllUsers = `-- name: DeleteAllUsers :exec
DELETE FROM users
`

func (q *Queries) DeleteAllUsers(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteAllUsers)
	return err
}

const getUserPassword = `-- name: GetUserPassword :one
SELECT id, created_at, updated_at, email, hashed_password, is_chirpy_red FROM users
WHERE email = $1
`

func (q *Queries) GetUserPassword(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserPassword, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
		&i.IsChirpyRed,
	)
	return i, err
}

const makeUserRed = `-- name: MakeUserRed :one
UPDATE users SET (is_chirpy_red, updated_at) = (true, NOW())
WHERE id = $1
RETURNING id, created_at, updated_at, email, hashed_password, is_chirpy_red
`

func (q *Queries) MakeUserRed(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRowContext(ctx, makeUserRed, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
		&i.IsChirpyRed,
	)
	return i, err
}

const updateUserDetails = `-- name: UpdateUserDetails :one
UPDATE users SET (email, hashed_password, updated_at) = ($1, $2, NOW())
WHERE id = $3
RETURNING id, created_at, updated_at, email, hashed_password, is_chirpy_red
`

type UpdateUserDetailsParams struct {
	Email          string
	HashedPassword string
	ID             uuid.UUID
}

func (q *Queries) UpdateUserDetails(ctx context.Context, arg UpdateUserDetailsParams) (User, error) {
	row := q.db.QueryRowContext(ctx, updateUserDetails, arg.Email, arg.HashedPassword, arg.ID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
		&i.IsChirpyRed,
	)
	return i, err
}
