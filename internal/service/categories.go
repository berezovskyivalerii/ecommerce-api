package service

import (
	"context"
	"fmt"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
)

type CategoriesDB interface {
	CreateCategory(ctx context.Context, name string) (postgres.Category, error)
	GetCategories(ctx context.Context) ([]postgres.Category, error)
	UpdateCategory(ctx context.Context, arg postgres.UpdateCategoryParams) (postgres.Category, error)
	GetCategoryByID(ctx context.Context, id int32) (postgres.Category, error)
	DeleteCategory(ctx context.Context, id int32) error
}

type CategoriesService struct {
	db CategoriesDB
}

func NewCategoriesService(
	db CategoriesDB,
) *CategoriesService {
	return &CategoriesService{
		db: db,
	}
}

func (s *CategoriesService) CreateCategory(ctx context.Context, name string) (models.Category, error) {
	if name == "" {
		return models.Category{}, fmt.Errorf("name is required")
	}

	category, err := s.db.CreateCategory(ctx, name)
	if err != nil {
		return models.Category{}, fmt.Errorf("failed to create category: %v", err)
	}

	return models.Category{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}, nil
}

func (s *CategoriesService) GetCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := s.db.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve categories: %v", err)
	}

	categories := make([]models.Category, 0, len(rows))
	for _, row := range rows {
		categories = append(categories, models.Category{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			Name:      row.Name,
		})
	}

	return categories, nil
}

func (s *CategoriesService) GetCategoryByID(ctx context.Context, id int32) (postgres.Category, error) {
	category, err := s.db.GetCategoryByID(ctx, id)
	if err != nil {
		return postgres.Category{}, fmt.Errorf("error retrieving category: %v", err)
	}
	return category, nil
}

func (s *CategoriesService) UpdateCategory(ctx context.Context, name string, id int32) (models.Category, error) {
	if name == "" || id < 1 {
		return models.Category{}, fmt.Errorf("invalid params")
	}

	category, err := s.db.UpdateCategory(ctx, postgres.UpdateCategoryParams{
		Name: name,
		ID:   id,
	})
	if err != nil {
		return models.Category{}, fmt.Errorf("failed to update category: %v", err)
	}

	return models.Category{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}, nil
}

func (s *CategoriesService) DeleteCategory(ctx context.Context, id int32) error {
	if id < 1 {
		return fmt.Errorf("id must be more than 1")
	}

	if err := s.db.DeleteCategory(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %v", err)
	}

	return nil
}
