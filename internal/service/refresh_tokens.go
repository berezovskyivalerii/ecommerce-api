package service

import (
	"context"
	"fmt"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
)

type RefreshTokensDB interface {
	CreateRefreshToken(ctx context.Context, arg postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error)
	GetRefreshToken(ctx context.Context, token string) (postgres.RefreshToken, error)
	RevokeToken(ctx context.Context, token string) error
}

type RefreshTokensService struct {
	db RefreshTokensDB
}

func NewRefreshTokensService(db RefreshTokensDB) *RefreshTokensService {
	return &RefreshTokensService{db: db}
}

func (s *RefreshTokensService) Create(ctx context.Context, req models.RefreshToken) (models.RefreshToken, error) {
	if req.Token == "" || req.UserID.String() == "" {
		return models.RefreshToken{}, fmt.Errorf("token and userID are required")
	}

	token, err := s.db.CreateRefreshToken(ctx, postgres.CreateRefreshTokenParams{
		Token:     req.Token,
		UserID:    req.UserID,
		ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		return models.RefreshToken{}, fmt.Errorf("failed to create refresh token: %v", err)
	}

	return models.RefreshToken{
		Token:     token.Token,
		CreatedAt: token.CreatedAt,
		UpdatedAt: token.UpdatedAt,
		UserID:    token.UserID,
		ExpiresAt: token.ExpiresAt,
		RevokedAt: &token.RevokedAt.Time,
	}, nil
}
