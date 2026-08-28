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

-- name: UpdateProduct :one
UPDATE products
SET name=$1, price_usd=$2, quantity=$3, category_id=$4, updated_at=NOW()
WHERE id=$5
RETURNING id, name, created_at, updated_at, price_usd, quantity, category_id;

-- name: GetProductsByCategoryID :many
SELECT id, name, created_at, updated_at, price_usd, quantity, category_id
FROM products
WHERE category_id=$1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountProductsByCategoryID :one
SELECT COUNT(*) FROM products WHERE category_id=$1;

-- name: SearchProducts :many
SELECT id, name, created_at, updated_at, price_usd, quantity, category_id
FROM products
WHERE name ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSearchProducts :one
SELECT COUNT(*) FROM products WHERE name ILIKE '%' || $1 || '%';
