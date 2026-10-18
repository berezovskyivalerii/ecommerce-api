package postgres

import (
	"context"
	"database/sql"
)

type ProductsRepository struct {
	store Store
}

func NewProductsRepository(store Store) *ProductsRepository {
	return &ProductsRepository{store: store}
}

func (r *ProductsRepository) CreateProduct(ctx context.Context, arg CreateProductParams) (CreateProductRow, error) {
	return r.store.CreateProduct(ctx, arg)
}

func (r *ProductsRepository) GetProductByID(ctx context.Context, id int32) (GetProductByIDRow, error) {
	return r.store.GetProductByID(ctx, id)
}

func (r *ProductsRepository) GetProducts(ctx context.Context, arg GetProductsParams) ([]GetProductsRow, error) {
	return r.store.GetProducts(ctx, arg)
}

func (r *ProductsRepository) CountProducts(ctx context.Context) (int64, error) {
	return r.store.CountProducts(ctx)
}

func (r *ProductsRepository) DeleteProduct(ctx context.Context, id int32) error {
	return r.store.DeleteProduct(ctx, id)
}

func (r *ProductsRepository) CountSearchProducts(ctx context.Context, websearchToTsquery string) (int64, error) {
	return r.store.CountSearchProducts(ctx, websearchToTsquery)
}

func (r *ProductsRepository) GetProductsByCategoryID(ctx context.Context, arg GetProductsByCategoryIDParams) ([]GetProductsByCategoryIDRow, error) {
	return r.store.GetProductsByCategoryID(ctx, arg)
}

func (r *ProductsRepository) CountProductsByCategoryID(ctx context.Context, categoryID sql.NullInt32) (int64, error) {
	return r.store.CountProductsByCategoryID(ctx, categoryID)
}

func (r *ProductsRepository) UpdateProduct(ctx context.Context, arg UpdateProductParams) (UpdateProductRow, error) {
	return r.store.UpdateProduct(ctx, arg)
}

func (r *ProductsRepository) SearchProducts(ctx context.Context, arg SearchProductsParams) ([]SearchProductsRow, error) {
	return r.store.SearchProducts(ctx, arg)
}
