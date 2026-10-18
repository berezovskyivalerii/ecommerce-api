package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	hasher "github.com/berezovskyivalerii/ecommerce-api/internal/pkg/hash"
	jwtmanager "github.com/berezovskyivalerii/ecommerce-api/internal/pkg/jwt"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
	"github.com/berezovskyivalerii/ecommerce-api/internal/service"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	_ = godotenv.Load("../../../.env")

	candidates := []string{}
	if v := os.Getenv("TEST_DB_URL"); v != "" {
		candidates = append(candidates, v)
	}
	if v := os.Getenv("LOCAL_DB_URL"); v != "" {
		candidates = append(candidates, v)
	}
	candidates = append(candidates, "postgres://postgres:postgres@localhost:5433/ecommerce?sslmode=disable")
	if v := os.Getenv("DB_URL"); v != "" {
		candidates = append(candidates, v)
	}

	var lastErr error
	for _, dsn := range candidates {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			lastErr = err
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			t.Cleanup(func() {
				_ = db.Close()
			})
			return db
		}
		_ = db.Close()
		lastErr = err
	}

	t.Skipf("skipping integration test: postgres is not reachable: %v", lastErr)
	return nil
}

func cleanupUserByEmail(t *testing.T, db *sql.DB, email string) {
	t.Helper()
	_, err := db.Exec(`DELETE FROM users WHERE email = $1`, email)
	if err != nil {
		t.Logf("warning: failed to cleanup test user %s: %v", email, err)
	}
}

func uniqueTestEmail(prefix string) string {
	return prefix + "_" + uuid.NewString()[:8] + "@example.com"
}

func TestUserRepository_Integration_CRUD(t *testing.T) {
	db := setupTestDB(t)
	store := postgres.NewStore(db)
	repo := postgres.NewUserRepository(store)
	ctx := context.Background()

	email := uniqueTestEmail("repo_crud")
	t.Cleanup(func() { cleanupUserByEmail(t, db, email) })

	// 1. CreateUser
	created, err := repo.CreateUser(ctx, postgres.CreateUserParams{
		Email:          email,
		HashedPassword: "hashed-password-123",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Errorf("expected non-nil UUID, got %v", created.ID)
	}
	if created.Email != email {
		t.Errorf("expected email %q, got %q", email, created.Email)
	}
	if created.Role != postgres.UserRoleUser {
		t.Errorf("expected default role %q, got %q", postgres.UserRoleUser, created.Role)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Errorf("expected timestamps to be set, got CreatedAt=%v, UpdatedAt=%v", created.CreatedAt, created.UpdatedAt)
	}

	// 2. GetUserByID
	byID, err := repo.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if byID.ID != created.ID || byID.Email != email || byID.HashedPassword != "hashed-password-123" {
		t.Errorf("GetUserByID mismatch: got %+v", byID)
	}

	// 3. GetUserByEmail
	byEmail, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if byEmail.ID != created.ID || byEmail.Role != postgres.UserRoleUser || byEmail.HashedPassword != "hashed-password-123" {
		t.Errorf("GetUserByEmail mismatch: got %+v", byEmail)
	}

	// 4. UpdateUser
	updatedEmail := uniqueTestEmail("repo_updated")
	t.Cleanup(func() { cleanupUserByEmail(t, db, updatedEmail) })

	updated, err := repo.UpdateUser(ctx, postgres.UpdateUserParams{
		Email:          updatedEmail,
		HashedPassword: "new-hashed-password-456",
		ID:             created.ID,
	})
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}
	if updated.Email != updatedEmail {
		t.Errorf("expected updated email %q, got %q", updatedEmail, updated.Email)
	}
	if updated.UpdatedAt.Before(created.UpdatedAt) {
		t.Errorf("expected UpdatedAt (%v) >= CreatedAt (%v)", updated.UpdatedAt, created.UpdatedAt)
	}

	// 5. CountUsers & GetUsers
	countBeforeDelete, err := repo.CountUsers(ctx)
	if err != nil {
		t.Fatalf("CountUsers failed: %v", err)
	}
	if countBeforeDelete < 1 {
		t.Errorf("expected CountUsers >= 1, got %d", countBeforeDelete)
	}

	usersList, err := repo.GetUsers(ctx, postgres.GetUsersParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}
	if len(usersList) == 0 {
		t.Errorf("expected non-empty GetUsers slice")
	}

	// 6. DeleteUser
	if err := repo.DeleteUser(ctx, created.ID); err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	_, err = repo.GetUserByID(ctx, created.ID)
	if err == nil {
		t.Errorf("expected error when getting deleted user, got nil")
	}
}

func TestUserRepository_Integration_ConstraintsAndAdmin(t *testing.T) {
	db := setupTestDB(t)
	store := postgres.NewStore(db)
	repo := postgres.NewUserRepository(store)
	ctx := context.Background()

	email := uniqueTestEmail("constraint_user")
	t.Cleanup(func() { cleanupUserByEmail(t, db, email) })

	// 1. Unique email constraint
	_, err := repo.CreateUser(ctx, postgres.CreateUserParams{
		Email:          email,
		HashedPassword: "hash1",
	})
	if err != nil {
		t.Fatalf("first CreateUser failed: %v", err)
	}

	_, err = repo.CreateUser(ctx, postgres.CreateUserParams{
		Email:          email,
		HashedPassword: "hash2",
	})
	if err == nil {
		t.Errorf("expected duplicate email error from PostgreSQL, got nil")
	}

	// 2. Email length CHECK constraint (VARCHAR(100) / char_length BETWEEN 1 AND 100)
	_, err = repo.CreateUser(ctx, postgres.CreateUserParams{
		Email:          "",
		HashedPassword: "hash",
	})
	if err == nil {
		t.Errorf("expected error for empty email violating CHECK constraint, got nil")
	}

	longEmail := strings.Repeat("a", 95) + "@example.com" // > 100 chars
	_, err = repo.CreateUser(ctx, postgres.CreateUserParams{
		Email:          longEmail,
		HashedPassword: "hash",
	})
	if err == nil {
		t.Errorf("expected error for email > 100 chars, got nil")
	}

	// 3. CreateAdmin
	adminEmail := uniqueTestEmail("admin_user")
	t.Cleanup(func() { cleanupUserByEmail(t, db, adminEmail) })

	adminRow, err := repo.CreateAdmin(ctx, postgres.CreateAdminParams{
		Email:          adminEmail,
		HashedPassword: "admin-hash",
	})
	if err != nil {
		t.Fatalf("CreateAdmin failed: %v", err)
	}
	if adminRow.Role != postgres.UserRoleAdmin {
		t.Errorf("expected admin role %q, got %q", postgres.UserRoleAdmin, adminRow.Role)
	}
}

func TestUserService_Integration_RegisterAndLifecycle(t *testing.T) {
	db := setupTestDB(t)
	store := postgres.NewStore(db)
	userRepo := postgres.NewUserRepository(store)
	refreshRepo := postgres.NewRefreshTokensRepository(store)
	passwordHasher := hasher.New()
	jwtMgr := jwtmanager.New()

	userService := service.NewUserService(userRepo, passwordHasher, jwtMgr, refreshRepo)
	ctx := context.Background()

	email := uniqueTestEmail("svc_reg")
	rawPassword := "StrongPass!234"
	t.Cleanup(func() { cleanupUserByEmail(t, db, email) })

	// 1. Register new user
	registeredUser, err := userService.Register(ctx, models.User{Email: email}, rawPassword)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if registeredUser.ID == uuid.Nil {
		t.Errorf("expected valid UUID after registration")
	}
	if registeredUser.Email != email {
		t.Errorf("expected email %q, got %q", email, registeredUser.Email)
	}
	if registeredUser.Role != "user" {
		t.Errorf("expected role 'user', got %q", registeredUser.Role)
	}
	if registeredUser.HashedPassword != "" {
		t.Errorf("expected HashedPassword not to be exposed in returned domain model, got %q", registeredUser.HashedPassword)
	}

	// 2. Verify stored password hash in DB is Argon2id and matches rawPassword
	dbUser, err := userRepo.GetUserByID(ctx, registeredUser.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if dbUser.HashedPassword == rawPassword {
		t.Fatalf("password was stored in plaintext!")
	}
	if err := passwordHasher.CheckPasswordHash(rawPassword, dbUser.HashedPassword); err != nil {
		t.Errorf("stored password hash did not match raw password: %v", err)
	}

	// 3. Duplicate registration should fail
	_, err = userService.Register(ctx, models.User{Email: email}, rawPassword)
	if err == nil {
		t.Errorf("expected error when registering duplicate email, got nil")
	}

	// 4. Registration validation (empty email / empty password)
	if _, err := userService.Register(ctx, models.User{Email: ""}, rawPassword); err == nil {
		t.Errorf("expected error for empty email")
	}
	if _, err := userService.Register(ctx, models.User{Email: uniqueTestEmail("no_pass")}, ""); err == nil {
		t.Errorf("expected error for empty password")
	}

	// 5. Login with registered user
	tokens, err := userService.Login(ctx, email, rawPassword)
	if err != nil {
		t.Fatalf("Login after registration failed: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Errorf("expected non-empty token pair, got %+v", tokens)
	}
}
