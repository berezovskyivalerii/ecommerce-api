-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING id, created_at, updated_at, email, role;

-- name: GetUsers :many
SELECT id, created_at, updated_at, email, role
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, hashed_password, role
FROM users
WHERE email=$1;

-- name: GetUserByID :one
SELECT id, created_at, updated_at, hashed_password, email, role
FROM users
WHERE id=$1;

-- name: UpdateUser :one
UPDATE users
SET email=$1, hashed_password=$2, updated_at=NOW()
WHERE id=$3
RETURNING id, created_at, updated_at, email, role;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id=$1;

-- name: CreateAdmin :one
INSERT INTO users (id, created_at,  updated_at, email, hashed_password, role)
VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2,
  'admin'
)
RETURNING id, created_at, updated_at, email, role;

