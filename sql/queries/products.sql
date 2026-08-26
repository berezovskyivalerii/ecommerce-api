-- name: CreateProduct :one
INSERT INTO products(name, created_at, updated_at, price_usd, quantity, category_id)
VALUES(
  $1,
  NOW(),
  NOW(),
  $2,
  $3,
  $4
) RETURNING id, name, created_at, updated_at, price_usd, quantity, category_id;

-- name: GetProducts :many
SELECT id, name, created_at, updated_at, price_usd, quantity, category_id
FROM products
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountProducts :one
SELECT COUNT(*) FROM products;

-- name: GetProductByID :one
SELECT id, name, created_at, updated_at, price_usd, quantity, category_id
FROM products
WHERE id=$1;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id=$1;
