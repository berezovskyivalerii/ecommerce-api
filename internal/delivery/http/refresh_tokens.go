package httprest

import (
	"context"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/google/uuid"
)

type RefreshTokensService interface {
	Create(ctx context.Context, req models.RefreshToken) (models.RefreshToken, error)
	Check(ctx context.Context, token string) (uuid.UUID, error)
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
