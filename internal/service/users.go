package service

import (
	"context"
	"fmt"
	"time"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
	"github.com/google/uuid"
)

type UserDB interface {
	CreateUser(ctx context.Context, arg postgres.CreateUserParams) (postgres.CreateUserRow, error)
	GetUsers(ctx context.Context) ([]postgres.GetUsersRow, error)
	GetUserByEmail(ctx context.Context, email string) (postgres.GetUserByEmailRow, error)
}

type HashManager interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) error
}

type JWTManager interface {
	MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error)
	MakeRefreshToken() string
}

type RefreshTokenSaver interface {
	CreateRefreshToken(ctx context.Context, req postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error)
}

type UserService struct {
	db          UserDB
	hashManager HashManager
	jwtManager  JWTManager
	tokenSaver  RefreshTokenSaver
}

func NewUserService(
	db UserDB,
	hashManager HashManager,
	jwtManager JWTManager,
	tokenSaver RefreshTokenSaver,
) *UserService {
	return &UserService{
		db:          db,
		hashManager: hashManager,
		jwtManager:  jwtManager,
		tokenSaver:  tokenSaver,
	}
}

func (s *UserService) Register(ctx context.Context, user models.User, password string) (models.User, error) {
	if password == "" || user.Email == "" {
		return models.User{}, fmt.Errorf("email and password are required")
	}

	hashedPassword, err := s.hashManager.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	createdUser, err := s.db.CreateUser(ctx, postgres.CreateUserParams{
		Email:          user.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		return models.User{}, err
	}

	return models.User{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}, nil
}

// Login authenticates user and returns an access token
func (s *UserService) Login(ctx context.Context, email, password string) (models.TokenPair, error) {
	if email == "" || password == "" {
		return models.TokenPair{}, fmt.Errorf("email and password are required")
	}

	userRow, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return models.TokenPair{}, fmt.Errorf("invalid credentials")
	}

	err = s.hashManager.CheckPasswordHash(password, userRow.HashedPassword)
	if err != nil {
		return models.TokenPair{}, fmt.Errorf("invalid credentials")
	}

	accessToken, err := s.jwtManager.MakeJWT(userRow.ID, "secret-123", time.Hour)
	if err != nil {
		return models.TokenPair{}, fmt.Errorf("failed to generate token: %v", err)
	}

	refreshToken := s.jwtManager.MakeRefreshToken()
	expiresAt := time.Now().Add(60 * 24 * time.Hour)

	_, err = s.tokenSaver.CreateRefreshToken(ctx, postgres.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    userRow.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return models.TokenPair{}, fmt.Errorf("failed to save refresh token: %v", err)
	}

	return models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
