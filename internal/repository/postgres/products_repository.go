package postgres

import "context"

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
