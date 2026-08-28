package postgres

import (
	"context"
	"database/sql"
)

type ProductRepository struct {
	store Store
}

func NewProductsRepository(store Store) *ProductRepository {
	return &ProductRepository{
		store: store,
	}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, arg CreateProductParams) (Product, error) {
	return r.store.CreateProduct(ctx, arg)
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id int32) (Product, error) {
	return r.store.GetProductByID(ctx, id)
}

func (r *ProductRepository) GetProducts(ctx context.Context, arg GetProductsParams) ([]Product, error) {
	return r.store.GetProducts(ctx, arg)
}

func (r *ProductRepository) CountProducts(ctx context.Context) (int64, error) {
	return r.store.CountProducts(ctx)
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id int32) error {
	return r.store.DeleteProduct(ctx, id)
}

func (r *ProductRepository) CountProductsByCategoryID(ctx context.Context, categoryID sql.NullInt32) (int64, error) {
	return r.store.CountProductsByCategoryID(ctx, categoryID)
}

func (r *ProductRepository) CountSearchProducts(ctx context.Context, dollar_1 sql.NullString) (int64, error) {
	return r.store.CountSearchProducts(ctx, dollar_1)
}

func (r *ProductRepository) GetProductsByCategoryID(ctx context.Context, arg GetProductsByCategoryIDParams) ([]Product, error) {
	return r.store.GetProductsByCategoryID(ctx, arg)
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, arg UpdateProductParams) (Product, error) {
	return r.store.UpdateProduct(ctx, arg)
}

func (r *ProductRepository) SearchProducts(ctx context.Context, arg SearchProductsParams) ([]Product, error) {
	return r.store.SearchProducts(ctx, arg)
}
