package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
	"github.com/google/uuid"
)

type UserDB interface {
	CreateUser(ctx context.Context, arg postgres.CreateUserParams) (postgres.CreateUserRow, error)
	CreateAdmin(ctx context.Context, arg postgres.CreateAdminParams) (postgres.CreateAdminRow, error)
	CountUsers(ctx context.Context) (int64, error)
	GetUsers(ctx context.Context, arg postgres.GetUsersParams) ([]postgres.GetUsersRow, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (postgres.GetUserByIDRow, error)
	GetUserByEmail(ctx context.Context, email string) (postgres.GetUserByEmailRow, error)
}

type HashManager interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) error
}

type JWTManager interface {
	MakeJWT(userID uuid.UUID, role string, tokenSecret string, expiresIn time.Duration) (string, error)
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

// Register create user in database
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
		Role:      string(createdUser.Role),
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

	accessToken, err := s.jwtManager.MakeJWT(userRow.ID, string(userRow.Role), "secret-123", time.Hour)
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

// GetUsers returns paginated list of users
func (s *UserService) GetUsers(ctx context.Context, page, limit int) (models.PaginatedUsers, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	totalCount, err := s.db.CountUsers(ctx)
	if err != nil {
		return models.PaginatedUsers{}, fmt.Errorf("failed to count users: %v", err)
	}

	rows, err := s.db.GetUsers(ctx, postgres.GetUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return models.PaginatedUsers{}, fmt.Errorf("failed to retrieve users: %v", err)
	}

	users := make([]models.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, models.User{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			Email:     row.Email,
			Role:      string(row.Role),
		})
	}

	totalPages := int(totalCount) / limit
	if int(totalCount)%limit != 0 {
		totalPages++
	}

	return models.PaginatedUsers{
		Data:       users,
		TotalCount: int(totalCount),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// SeedAdmin create user with admin role
func SeedAdmin(ctx context.Context, db UserDB, hashManager HashManager, adminEmail, adminPassword string) {
	if adminEmail == "" || adminPassword == "" {
		log.Println("admin seeding skipped: credentials not provided in environment")
		return
	}

	_, err := db.GetUserByEmail(ctx, adminEmail)
	if err == nil {
		log.Println("admin seeding skipped: email already registered")
		return
	}

	hashedPassword, err := hashManager.HashPassword(adminPassword)
	if err != nil {
		log.Fatalf("failed to hash admin's password: %v", err)
	}

	_, err = db.CreateAdmin(ctx, postgres.CreateAdminParams{
		Email:          adminEmail,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Fatalf("failed to seed admin user: %v", err)
	}

	log.Println("admin was successfully seeded")
}
