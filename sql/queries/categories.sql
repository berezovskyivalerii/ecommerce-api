-- name: CreateCategory :one
INSERT INTO categories (name, created_at, updated_at) 
VALUES (
  $1,
  NOW(),
  NOW()
)
RETURNING id, name, created_at, updated_at;

-- name: GetCategories :many
SELECT id, name, created_at, updated_at
FROM categories;

-- name: UpdateCategory :one
UPDATE categories
SET name=$1, updated_at=NOW()
WHERE id=$2
RETURNING id, name, created_at, updated_at;

-- name: DeleteCategory :exec
DELETE
FROM categories
WHERE id=$1;
