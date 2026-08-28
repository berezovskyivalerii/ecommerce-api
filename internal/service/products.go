package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
)

type ProductsDB interface {
	CreateProduct(ctx context.Context, arg postgres.CreateProductParams) (postgres.Product, error)
	GetProductByID(ctx context.Context, id int32) (postgres.Product, error)
	GetProducts(ctx context.Context, arg postgres.GetProductsParams) ([]postgres.Product, error)
	CountProducts(ctx context.Context) (int64, error)
	DeleteProduct(ctx context.Context, id int32) error

	UpdateProduct(ctx context.Context, arg postgres.UpdateProductParams) (postgres.Product, error)
	GetProductsByCategoryID(ctx context.Context, arg postgres.GetProductsByCategoryIDParams) ([]postgres.Product, error)
	CountProductsByCategoryID(ctx context.Context, categoryID sql.NullInt32) (int64, error)
	SearchProducts(ctx context.Context, arg postgres.SearchProductsParams) ([]postgres.Product, error)
	CountSearchProducts(ctx context.Context, dollar_1 sql.NullString) (int64, error)
}

type CategoriesManager interface {
	GetCategoryByID(ctx context.Context, id int32) (postgres.Category, error)
}

type ProductsService struct {
	db         ProductsDB
	categories CategoriesManager
}

func NewProductsService(db ProductsDB, categories CategoriesManager) *ProductsService {
	return &ProductsService{db: db, categories: categories}
}

func (s *ProductsService) CreateProduct(ctx context.Context, product models.Product) (models.Product, error) {
	if _, err := s.categories.GetCategoryByID(ctx, int32(product.CategotyID)); err != nil {
		return models.Product{}, fmt.Errorf("error: no category with such ID: %v", err)
	}

	productCreated, err := s.db.CreateProduct(ctx, postgres.CreateProductParams{
		Name:       product.Name,
		PriceUsd:   strconv.FormatFloat(product.PriceUSD, 'f', -1, 64),
		Quantity:   int32(product.Quantity),
		CategoryID: sql.NullInt32{Int32: int32(product.CategotyID), Valid: true},
	})
	if err != nil {
		return models.Product{}, fmt.Errorf("error: cannot create product: %v", err)
	}
	priceUSD, _ := strconv.ParseFloat(productCreated.PriceUsd, 8)
	return models.Product{
		ID:         productCreated.ID,
		Name:       productCreated.Name,
		PriceUSD:   priceUSD,
		Quantity:   productCreated.Quantity,
		CreatedAt:  productCreated.CreatedAt,
		UpdatedAt:  productCreated.UpdatedAt,
		CategotyID: productCreated.CategoryID.Int32,
	}, nil
}

func (s *ProductsService) GetProducts(ctx context.Context, page, limit int) (models.PaginatedProducts, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	totalCount, err := s.db.CountProducts(ctx)
	if err != nil {
		return models.PaginatedProducts{}, fmt.Errorf("failed to count products: %v", err)
	}

	rows, err := s.db.GetProducts(ctx, postgres.GetProductsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return models.PaginatedProducts{}, fmt.Errorf("failed to retrieve products: %v", err)
	}

	products := make([]models.Product, 0, len(rows))
	for _, row := range rows {
		priceUSD, _ := strconv.ParseFloat(row.PriceUsd, 8)
		products = append(products, models.Product{
			ID:         row.ID,
			Name:       row.Name,
			PriceUSD:   priceUSD,
			Quantity:   row.Quantity,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
			CategotyID: row.CategoryID.Int32,
		})
	}

	totalPages := int(totalCount) / limit
	if int(totalCount)%limit != 0 {
		totalPages++
	}

	return models.PaginatedProducts{
		Data:       products,
		TotalCount: int(totalCount),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *ProductsService) DeleteProduct(ctx context.Context, id int32) error {
	if id < 0 {
		return fmt.Errorf("invalid id format: ID must be more than 0")
	}
	if err := s.db.DeleteProduct(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product: %v", err)
	}
	return nil
}

func (s *ProductsService) UpdateProduct(ctx context.Context, id int32, product models.Product) (models.Product, error) {
	if _, err := s.categories.GetCategoryByID(ctx, int32(product.CategotyID)); err != nil {
		return models.Product{}, fmt.Errorf("error: no category with such ID: %v", err)
	}

	productUpdated, err := s.db.UpdateProduct(ctx, postgres.UpdateProductParams{
		ID:         id,
		Name:       product.Name,
		PriceUsd:   strconv.FormatFloat(product.PriceUSD, 'f', -1, 64),
		Quantity:   int32(product.Quantity),
		CategoryID: sql.NullInt32{Int32: int32(product.CategotyID), Valid: true},
	})
	if err != nil {
		return models.Product{}, fmt.Errorf("error: cannot update product: %v", err)
	}

	priceUSD, _ := strconv.ParseFloat(productUpdated.PriceUsd, 8)
	return models.Product{
		ID:         productUpdated.ID,
		Name:       productUpdated.Name,
		PriceUSD:   priceUSD,
		Quantity:   productUpdated.Quantity,
		CreatedAt:  productUpdated.CreatedAt,
		UpdatedAt:  productUpdated.UpdatedAt,
		CategotyID: productUpdated.CategoryID.Int32,
	}, nil
}

func (s *ProductsService) GetProductsByCategory(ctx context.Context, categoryID int32, page, limit int) (models.PaginatedProducts, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit
	catID := sql.NullInt32{Int32: categoryID, Valid: true}

	totalCount, err := s.db.CountProductsByCategoryID(ctx, catID)
	if err != nil {
		return models.PaginatedProducts{}, fmt.Errorf("failed to count products by category: %v", err)
	}

	rows, err := s.db.GetProductsByCategoryID(ctx, postgres.GetProductsByCategoryIDParams{
		CategoryID: catID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return models.PaginatedProducts{}, fmt.Errorf("failed to retrieve products by category: %v", err)
	}

	products := make([]models.Product, 0, len(rows))
	for _, row := range rows {
		priceUSD, _ := strconv.ParseFloat(row.PriceUsd, 8)
		products = append(products, models.Product{
			ID:         row.ID,
			Name:       row.Name,
			PriceUSD:   priceUSD,
			Quantity:   row.Quantity,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
			CategotyID: row.CategoryID.Int32,
		})
	}

	totalPages := int(totalCount) / limit
	if int(totalCount)%limit != 0 {
		totalPages++
	}

	return models.PaginatedProducts{
		Data:       products,
		TotalCount: int(totalCount),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *ProductsService) SearchProducts(ctx context.Context, query string, page, limit int) (models.PaginatedProducts, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit
	searchQuery := sql.NullString{String: query, Valid: true}

	totalCount, err := s.db.CountSearchProducts(ctx, searchQuery)
	if err != nil {
		return models.PaginatedProducts{}, fmt.Errorf("failed to count searched products: %v", err)
	}

	rows, err := s.db.SearchProducts(ctx, postgres.SearchProductsParams{
		Column1: searchQuery,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return models.PaginatedProducts{}, fmt.Errorf("failed to search products: %v", err)
	}

	products := make([]models.Product, 0, len(rows))
	for _, row := range rows {
		priceUSD, _ := strconv.ParseFloat(row.PriceUsd, 8)
		products = append(products, models.Product{
			ID:         row.ID,
			Name:       row.Name,
			PriceUSD:   priceUSD,
			Quantity:   row.Quantity,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
			CategotyID: row.CategoryID.Int32,
		})
	}

	totalPages := int(totalCount) / limit
	if int(totalCount)%limit != 0 {
		totalPages++
	}

	return models.PaginatedProducts{
		Data:       products,
		TotalCount: int(totalCount),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
