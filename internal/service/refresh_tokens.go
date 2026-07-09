package service

import (
	"context"
	"fmt"
	"time"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
)

type RefreshTokensDB interface {
	CreateRefreshToken(ctx context.Context, arg postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error)
	GetRefreshToken(ctx context.Context, token string) (postgres.RefreshToken, error)
	RevokeToken(ctx context.Context, token string) error
}

type RefreshTokensService struct {
	db         RefreshTokensDB
	jwtManager JWTManager
}

func NewRefreshTokensService(
	db RefreshTokensDB,
	jwtManager JWTManager,
) *RefreshTokensService {
	return &RefreshTokensService{
		db:         db,
		jwtManager: jwtManager,
	}
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

func (s *RefreshTokensService) Refresh(ctx context.Context, token string) (models.TokenPair, error) {
	if token == "" {
		return models.TokenPair{}, fmt.Errorf("token is empty")
	}

	// 1. Get refresh token from database
	refreshToken, err := s.db.GetRefreshToken(ctx, token)
	if err != nil {
		return models.TokenPair{}, fmt.Errorf("error get refresh token fro DB: %v", err)
	}

	// 2. Check if it is correct
	if refreshToken.ExpiresAt.Before(time.Now()) {
		return models.TokenPair{}, fmt.Errorf("refresh token expired")
	}

	if refreshToken.RevokedAt.Valid {
		return models.TokenPair{}, fmt.Errorf("refresh token revoked")
	}

	// 3. Revoke old token for safe
	err = s.db.RevokeToken(ctx, token)
	if err != nil {
		return models.TokenPair{}, fmt.Errorf("failed to revoke old token: %v", err)
	}

	// 4. Create new access token
	accessToken, err := s.jwtManager.MakeJWT(refreshToken.UserID, "secret-123", time.Hour)
	if err != nil {
		return models.TokenPair{}, fmt.Errorf("error creating access token: %v", err)
	}

	// 5. Create new refresh token
	newRefreshTokenString := s.jwtManager.MakeRefreshToken()
	expiresAt := time.Now().Add(60 * 24 * time.Hour)

	_, err = s.Create(ctx, models.RefreshToken{
		Token:     newRefreshTokenString,
		UserID:    refreshToken.UserID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return models.TokenPair{}, fmt.Errorf("failed to create refresh token: %v", err)
	}

	return models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenString,
	}, nil
}

func (s *RefreshTokensService) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token is empty")
	}

	err := s.db.RevokeToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %v", err)
	}

	return nil
}
