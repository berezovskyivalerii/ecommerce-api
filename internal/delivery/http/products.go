package httprest

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/gin-gonic/gin"
)

type ProductsService interface {
	CreateProduct(ctx context.Context, product models.Product) (models.Product, error)
	GetProducts(ctx context.Context, page, limit int) (models.PaginatedProducts, error)
	DeleteProduct(ctx context.Context, id int32) error
	UpdateProduct(ctx context.Context, id int32, product models.Product) (models.Product, error)
	GetProductsByCategory(ctx context.Context, categoryID int32, page, limit int) (models.PaginatedProducts, error)
	SearchProducts(ctx context.Context, query string, page, limit int) (models.PaginatedProducts, error)
}

type ProductsHandlers struct {
	service ProductsService
}

func NewProductsHandlers(service ProductsService) *ProductsHandlers {
	return &ProductsHandlers{
		service: service,
	}
}

type CreateProductRequest struct {
	Name       string  `json:"name" binding:"required"`
	PriceUSD   float64 `json:"price_usd" binding:"required,gt=0"`
	Quantity   int32   `json:"quantity" binding:"required,gte=0"`
	CategoryID int32   `json:"category_id" binding:"required,gt=0"`
}

func (h *ProductsHandlers) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to decode request: %v", err),
		})
		return
	}

	product := models.Product{
		Name:       req.Name,
		PriceUSD:   req.PriceUSD,
		Quantity:   req.Quantity,
		CategotyID: req.CategoryID,
	}

	createdProduct, err := h.service.CreateProduct(c.Request.Context(), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create product: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, createdProduct)
}

func (h *ProductsHandlers) UpdateProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil || productID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid product id format",
		})
		return
	}

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to decode request: %v", err),
		})
		return
	}

	product := models.Product{
		Name:       req.Name,
		PriceUSD:   req.PriceUSD,
		Quantity:   req.Quantity,
		CategotyID: req.CategoryID,
	}

	updatedProduct, err := h.service.UpdateProduct(c.Request.Context(), int32(productID), product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to update product: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, updatedProduct)
}

func (h *ProductsHandlers) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, err := h.service.GetProducts(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get products: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *ProductsHandlers) GetProductsByCategory(c *gin.Context) {
	categoryIDStr := c.Param("category_id")
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil || categoryID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid category id format",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, err := h.service.GetProductsByCategory(c.Request.Context(), int32(categoryID), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to get products by category: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *ProductsHandlers) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "search query parameter 'q' is required",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, err := h.service.SearchProducts(c.Request.Context(), query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to search products: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *ProductsHandlers) DeleteProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil || productID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid product id format",
		})
		return
	}

	err = h.service.DeleteProduct(c.Request.Context(), int32(productID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to delete product: %v", err),
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
