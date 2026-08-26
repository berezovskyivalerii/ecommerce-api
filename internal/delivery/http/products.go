package httprest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/gin-gonic/gin"
)

type ProductsService interface {
	GetProducts(ctx context.Context, page, limit int) (models.PaginatedProducts, error)
	CreateProduct(ctx context.Context, product models.Product) (models.Product, error)
	DeleteProduct(ctx context.Context, id int32) error
}

type ProductsHandlers struct {
	service ProductsService
}

func NewProductsHandlers(service ProductsService) *ProductsHandlers {
	return &ProductsHandlers{service: service}
}

type CreateProductRequest struct {
	Name       string  `json:"name"`
	PriceUSD   float64 `json:"price_usd"`
	Quantity   int32   `json:"quantity"`
	CategotyID int32   `json:"category_id"`
}

type CreateProductResponse struct {
	ID         int32     `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	PriceUSD   float64   `json:"price_usd"`
	Quantity   int32     `json:"quantity"`
	CategotyID int32     `json:"category_id"`
}

func (h *ProductsHandlers) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to decode request: %v", err),
		})
		return
	}

	product := &models.Product{
		Name:       req.Name,
		PriceUSD:   req.PriceUSD,
		Quantity:   req.Quantity,
		CategotyID: req.CategotyID,
	}

	createdProduct, err := h.service.CreateProduct(c.Request.Context(), *product)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create product: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, CreateProductResponse{
		ID:         createdProduct.ID,
		Name:       createdProduct.Name,
		CreatedAt:  createdProduct.CreatedAt,
		UpdatedAt:  createdProduct.UpdatedAt,
		PriceUSD:   createdProduct.PriceUSD,
		Quantity:   createdProduct.Quantity,
		CategotyID: createdProduct.CategotyID,
	})
}

func (h *ProductsHandlers) GetProducts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	paginatedProducts, err := h.service.GetProducts(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to fetch products: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, paginatedProducts)
}

func (h *ProductsHandlers) DeleteProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	if productIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "product id required in path",
		})
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid product id format",
		})
		return
	}

	err = h.service.DeleteProduct(c.Request.Context(), int32(productID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("failed to delete product: %v", err),
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
