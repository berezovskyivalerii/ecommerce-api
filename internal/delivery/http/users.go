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
	"github.com/google/uuid"
)

type UserService interface {
	Register(ctx context.Context, user models.User, password string) (models.User, error)
	Login(ctx context.Context, email, password string) (models.TokenPair, error)
	GetUsers(ctx context.Context, page, limit int) (models.PaginatedUsers, error)
	UpdateUser(ctx context.Context, email, password string, userID uuid.UUID) (models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type UserHandlers struct {
	usersService UserService
}

func NewUserHandlers(userService UserService,
) *UserHandlers {
	return &UserHandlers{
		usersService: userService,
	}
}

type CreateUserRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type CreateUserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Role      string    `json:"role"`
}

func (h *UserHandlers) Register(c *gin.Context) {
	var req CreateUserRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to decode request: %v", err),
		})
		return
	}

	user := &models.User{
		Email: req.Email,
	}

	createdUser, err := h.usersService.Register(c.Request.Context(), *user, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("failed to create user: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, CreateUserResponse{
		ID:        createdUser.ID.String(),
		Email:     createdUser.Email,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Role:      createdUser.Role,
	})
}

type LoginUserRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (h *UserHandlers) Login(c *gin.Context) {
	var req LoginUserRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to decode request: %v", err),
		})
		return
	}

	tokenPair, err := h.usersService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": fmt.Sprintf("login failed: %v", err),
		})
		return
	}

	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(60*24*time.Hour),
		"/",
		"",
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenPair.AccessToken,
	})
}

func (h *UserHandlers) GetUsers(c *gin.Context) {
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

	paginatedUsers, err := h.usersService.GetUsers(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to fetch users: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, paginatedUsers)
}

type UpdateUserRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Role      string    `json:"role"`
}

func (h *UserHandlers) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user id required in path",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id format",
		})
		return
	}

	var req UpdateUserRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("failed to decode request: %v", err),
		})
		return
	}

	updatedUser, err := h.usersService.UpdateUser(c.Request.Context(), req.Email, req.Password, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to update user: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, UpdateUserResponse{
		ID:        updatedUser.ID.String(),
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Role:      updatedUser.Role,
	})
}

func (h *UserHandlers) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user id required in path",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id format",
		})
		return
	}

	err = h.usersService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("failed to delete user: %v", err),
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
