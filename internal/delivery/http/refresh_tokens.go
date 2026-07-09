package httprest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/gin-gonic/gin"
)

type RefreshTokensService interface {
	Refresh(ctx context.Context, token string) (models.TokenPair, error)
	Revoke(ctx context.Context, token string) error
}

type RefreshTokensHandlers struct {
	refreshTokensService RefreshTokensService
}

func NewRefreshTokensHandlers(
	refreshTokensService RefreshTokensService,
) *RefreshTokensHandlers {
	return &RefreshTokensHandlers{
		refreshTokensService: refreshTokensService,
	}
}

func (h *RefreshTokensHandlers) RefreshToken(c *gin.Context) {
	tokenString, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "refresh token not found",
		})
	}
	tokenPair, err := h.refreshTokensService.Refresh(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("error getting access token: %v", err),
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

func (h *RefreshTokensHandlers) RevokeToken(c *gin.Context) {
	tokenString, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "refresh token not found",
		})
		return
	}

	err = h.refreshTokensService.Revoke(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("error revoking refresh token: %v", err),
		})
		return
	}

	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{})
}
