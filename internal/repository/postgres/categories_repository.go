package postgres

import (
	"context"
)

// CategoriesRepository wraps the Store interface
type CategoriesRepository struct {
	store Store
}

// NewCategoryRepository initializes UserRepository
func NewCategoryRepository(store Store) *CategoriesRepository {
	return &CategoriesRepository{
		store: store,
	}
}

// CreateCategory executes a single query using the embedded Querier
func (r *CategoriesRepository) CreateCategory(ctx context.Context, name string) (Category, error) {
	return r.store.CreateCategory(ctx, name)
}

func (r *CategoriesRepository) GetCategories(ctx context.Context) ([]Category, error) {
	return r.store.GetCategories(ctx)
}

func (r *CategoriesRepository) UpdateCategory(ctx context.Context, arg UpdateCategoryParams) (Category, error) {
	return r.store.UpdateCategory(ctx, arg)
}

func (r *CategoriesRepository) DeleteCategory(ctx context.Context, id int32) error {
	return r.store.DeleteCategory(ctx, id)
}
