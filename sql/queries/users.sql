-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
  gen_random_uuid(),
  NOW(),
  NOW(),
  $1,
  $2
)
RETURNING id, created_at, updated_at, email;

-- name: GetUsers :many
SELECT id, created_at, updated_at, email
FROM users
ORDER BY created_at DESC;

-- name: GetUserByEmail :one
SELECT id, created_at, updated_at, hashed_password
FROM users
WHERE email=$1;

