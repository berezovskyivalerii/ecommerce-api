package httprest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/gin-gonic/gin"
)

type UserService interface {
	Register(ctx context.Context, user models.User, password string) (models.User, error)
	Login(ctx context.Context, email, password string) (models.TokenPair, error)
}

type UserHandlers struct {
	usersService UserService
}

func NewUserHandlers(
	userService UserService,
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
