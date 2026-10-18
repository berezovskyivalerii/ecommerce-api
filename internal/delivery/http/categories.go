package httprest

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
	"github.com/gin-gonic/gin"
)

type CategoriesService interface {
	CreateCategory(ctx context.Context, name string) (models.Category, error)
	UpdateCategory(ctx context.Context, name string, id int32) (models.Category, error)
	GetCategories(ctx context.Context) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, id int32) (postgres.Category, error)
	DeleteCategory(ctx context.Context, id int32) error
}

type CategoriesHandlers struct {
	service CategoriesService
}

func NewCategoriesHandlers(service CategoriesService) *CategoriesHandlers {
	return &CategoriesHandlers{
		service: service,
	}
}

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateCategoryResponse struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *CategoriesHandlers) CreateCategory(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to decode request: %v", err),
		})
		return
	}

	createdCategory, err := h.service.CreateCategory(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create category: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, CreateCategoryResponse{
		ID:        createdCategory.ID,
		Name:      createdCategory.Name,
		CreatedAt: createdCategory.CreatedAt,
		UpdatedAt: createdCategory.UpdatedAt,
	})
}

func (h *CategoriesHandlers) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get categories: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *CategoriesHandlers) GetCategoryByID(c *gin.Context) {
	categoryIDStr := c.Param("id")
	if categoryIDStr == "" {
		categoryIDStr = c.Param("category_id")
	}
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id format",
		})
		return
	}

	category, err := h.service.GetCategoryByID(c.Request.Context(), int32(categoryID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("category not found: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.Category{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	})
}

type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *CategoriesHandlers) UpdateCategory(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id format",
		})
		return
	}

	var req UpdateCategoryRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to decode request: %v", err),
		})
		return
	}

	updatedCategory, err := h.service.UpdateCategory(c.Request.Context(), req.Name, int32(categoryID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to update category: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, updatedCategory)
}

func (h *CategoriesHandlers) DeleteCategory(c *gin.Context) {
	categoryIDStr := c.Param("id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id format",
		})
		return
	}

	err = h.service.DeleteCategory(c.Request.Context(), int32(categoryID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to delete category: %v", err),
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
