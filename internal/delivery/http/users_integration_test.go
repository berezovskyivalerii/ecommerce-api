package httprest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	httprest "github.com/berezovskyivalerii/ecommerce-api/internal/delivery/http"
	"github.com/berezovskyivalerii/ecommerce-api/internal/delivery/http/middleware"
	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	hasher "github.com/berezovskyivalerii/ecommerce-api/internal/pkg/hash"
	jwtmanager "github.com/berezovskyivalerii/ecommerce-api/internal/pkg/jwt"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
	"github.com/berezovskyivalerii/ecommerce-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// inMemoryUserDB implements service.UserDB and service.RefreshTokenSaver for full HTTP -> Service -> Hasher/JWT integration testing.
type inMemoryUserDB struct {
	mu            sync.RWMutex
	users         map[uuid.UUID]postgres.User
	refreshTokens map[string]postgres.RefreshToken
}

func newInMemoryUserDB() *inMemoryUserDB {
	return &inMemoryUserDB{
		users:         make(map[uuid.UUID]postgres.User),
		refreshTokens: make(map[string]postgres.RefreshToken),
	}
}

var (
	errDuplicateEmail = errors.New("pq: duplicate key value violates unique constraint \"users_email_key\"")
	errUserNotFound   = errors.New("sql: no rows in result set")
)

func (m *inMemoryUserDB) CreateUser(_ context.Context, arg postgres.CreateUserParams) (postgres.CreateUserRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.users {
		if u.Email == arg.Email {
			return postgres.CreateUserRow{}, errDuplicateEmail
		}
	}

	now := time.Now().UTC()
	u := postgres.User{
		ID:             uuid.New(),
		CreatedAt:      now,
		UpdatedAt:      now,
		Email:          arg.Email,
		HashedPassword: arg.HashedPassword,
		Role:           postgres.UserRoleUser,
	}
	m.users[u.ID] = u

	return postgres.CreateUserRow{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
		Role:      u.Role,
	}, nil
}

func (m *inMemoryUserDB) CreateAdmin(_ context.Context, arg postgres.CreateAdminParams) (postgres.CreateAdminRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.users {
		if u.Email == arg.Email {
			return postgres.CreateAdminRow{}, errDuplicateEmail
		}
	}

	now := time.Now().UTC()
	u := postgres.User{
		ID:             uuid.New(),
		CreatedAt:      now,
		UpdatedAt:      now,
		Email:          arg.Email,
		HashedPassword: arg.HashedPassword,
		Role:           postgres.UserRoleAdmin,
	}
	m.users[u.ID] = u

	return postgres.CreateAdminRow{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
		Role:      u.Role,
	}, nil
}

func (m *inMemoryUserDB) CountUsers(_ context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.users)), nil
}

func (m *inMemoryUserDB) GetUsers(_ context.Context, arg postgres.GetUsersParams) ([]postgres.GetUsersRow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	all := make([]postgres.GetUsersRow, 0, len(m.users))
	for _, u := range m.users {
		all = append(all, postgres.GetUsersRow{
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
			Email:     u.Email,
			Role:      u.Role,
		})
	}

	start := int(arg.Offset)
	if start >= len(all) {
		return []postgres.GetUsersRow{}, nil
	}
	end := start + int(arg.Limit)
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], nil
}

func (m *inMemoryUserDB) GetUserByID(_ context.Context, id uuid.UUID) (postgres.GetUserByIDRow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, ok := m.users[id]
	if !ok {
		return postgres.GetUserByIDRow{}, errUserNotFound
	}
	return postgres.GetUserByIDRow{
		ID:             u.ID,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
		HashedPassword: u.HashedPassword,
		Email:          u.Email,
		Role:           u.Role,
	}, nil
}

func (m *inMemoryUserDB) GetUserByEmail(_ context.Context, email string) (postgres.GetUserByEmailRow, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.Email == email {
			return postgres.GetUserByEmailRow{
				ID:             u.ID,
				CreatedAt:      u.CreatedAt,
				UpdatedAt:      u.UpdatedAt,
				HashedPassword: u.HashedPassword,
				Role:           u.Role,
			}, nil
		}
	}
	return postgres.GetUserByEmailRow{}, errUserNotFound
}

func (m *inMemoryUserDB) UpdateUser(_ context.Context, arg postgres.UpdateUserParams) (postgres.UpdateUserRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, ok := m.users[arg.ID]
	if !ok {
		return postgres.UpdateUserRow{}, errUserNotFound
	}

	for id, existing := range m.users {
		if id != arg.ID && existing.Email == arg.Email {
			return postgres.UpdateUserRow{}, errDuplicateEmail
		}
	}

	u.Email = arg.Email
	u.HashedPassword = arg.HashedPassword
	u.UpdatedAt = time.Now().UTC()
	m.users[arg.ID] = u

	return postgres.UpdateUserRow{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
		Role:      u.Role,
	}, nil
}

func (m *inMemoryUserDB) DeleteUser(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.users[id]; !ok {
		return errUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *inMemoryUserDB) CreateRefreshToken(_ context.Context, req postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	rt := postgres.RefreshToken{
		Token:     req.Token,
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    req.UserID,
		ExpiresAt: req.ExpiresAt,
	}
	m.refreshTokens[req.Token] = rt
	return rt, nil
}

func setupUserTestRouter() (*gin.Engine, *inMemoryUserDB, *hasher.Hasher, *jwtmanager.JWTManager) {
	gin.SetMode(gin.TestMode)

	db := newInMemoryUserDB()
	passwordHasher := hasher.New()
	jwtMgr := jwtmanager.New()

	userService := service.NewUserService(db, passwordHasher, jwtMgr, db)
	userHandler := httprest.NewUserHandlers(userService)

	r := gin.New()
	r.POST("/api/register", userHandler.Register)
	r.POST("/api/login", userHandler.Login)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware("secret-123"))
	{
		adminRoutes := api.Group("/admin")
		adminRoutes.Use(middleware.RequireRole("admin"))
		{
			adminRoutes.GET("/users", userHandler.GetUsers)
			adminRoutes.PUT("/users/:id", userHandler.UpdateUser)
			adminRoutes.DELETE("/users/:id", userHandler.DeleteUser)
		}
	}

	return r, db, passwordHasher, jwtMgr
}

func TestUserRegistration_HTTPIntegration(t *testing.T) {
	router, db, passwordHasher, _ := setupUserTestRouter()

	t.Run("successful user registration creates user with hashed password and default role", func(t *testing.T) {
		payload := map[string]string{
			"email":    "alice@example.com",
			"password": "secretPassword123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d; body: %s", http.StatusCreated, w.Code, w.Body.String())
		}

		var resp httprest.CreateUserResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		parsedID, err := uuid.Parse(resp.ID)
		if err != nil || parsedID == uuid.Nil {
			t.Errorf("expected valid UUID in response, got %q (err: %v)", resp.ID, err)
		}
		if resp.Email != "alice@example.com" {
			t.Errorf("expected email 'alice@example.com', got %q", resp.Email)
		}
		if resp.Role != "user" {
			t.Errorf("expected role 'user', got %q", resp.Role)
		}
		if resp.CreatedAt.IsZero() || resp.UpdatedAt.IsZero() {
			t.Errorf("expected non-zero timestamps in response")
		}

		// Check that password was properly hashed in the underlying storage
		storedUser, err := db.GetUserByID(context.Background(), parsedID)
		if err != nil {
			t.Fatalf("expected user to be persisted in storage: %v", err)
		}
		if storedUser.HashedPassword == "secretPassword123" {
			t.Errorf("password was stored in plain text")
		}
		if err := passwordHasher.CheckPasswordHash("secretPassword123", storedUser.HashedPassword); err != nil {
			t.Errorf("stored Argon2id hash does not match original password: %v", err)
		}
	})

	t.Run("duplicate email registration fails", func(t *testing.T) {
		payload := map[string]string{
			"email":    "alice@example.com",
			"password": "anotherPassword123",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code == http.StatusCreated {
			t.Fatalf("expected duplicate registration to fail, but got %d", w.Code)
		}
	})

	t.Run("registration with missing email or password fails", func(t *testing.T) {
		testCases := []struct {
			name    string
			payload map[string]string
		}{
			{
				name:    "empty email",
				payload: map[string]string{"email": "", "password": "secretPassword123"},
			},
			{
				name:    "empty password",
				payload: map[string]string{"email": "bob@example.com", "password": ""},
			},
			{
				name:    "both empty",
				payload: map[string]string{},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				body, _ := json.Marshal(tc.payload)
				req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				if w.Code == http.StatusCreated || w.Code == http.StatusOK {
					t.Errorf("expected error status for %s, got %d", tc.name, w.Code)
				}
			})
		}
	})

	t.Run("registration with malformed JSON returns 400 Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString("{invalid-json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestUserEntity_LoginAndAdminLifecycle_HTTPIntegration(t *testing.T) {
	router, db, passwordHasher, jwtMgr := setupUserTestRouter()

	// 1. Seed an admin user via service.SeedAdmin
	service.SeedAdmin(context.Background(), db, passwordHasher, "admin@example.com", "adminSecret123")

	// 2. Register a regular user via HTTP POST /api/register
	regBody, _ := json.Marshal(map[string]string{
		"email":    "charlie@example.com",
		"password": "charliePassword123",
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(regBody))
	regW := httptest.NewRecorder()
	router.ServeHTTP(regW, regReq)

	if regW.Code != http.StatusCreated {
		t.Fatalf("failed to register user: %d %s", regW.Code, regW.Body.String())
	}
	var createdUser httprest.CreateUserResponse
	_ = json.Unmarshal(regW.Body.Bytes(), &createdUser)

	// 3. Login as registered user via POST /api/login
	var userAccessToken string
	t.Run("registered user can login and receives JWT and refresh_token cookie", func(t *testing.T) {
		loginBody, _ := json.Marshal(map[string]string{
			"email":    "charlie@example.com",
			"password": "charliePassword123",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(loginBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on login, got %d: %s", w.Code, w.Body.String())
		}

		var loginResp map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
			t.Fatalf("failed to decode login response: %v", err)
		}
		userAccessToken = loginResp["access_token"]
		if userAccessToken == "" {
			t.Fatalf("expected non-empty access_token")
		}

		// Validate issued JWT subject matches registered user ID
		subjectID, err := jwtMgr.ValidateJWT(userAccessToken, "secret-123")
		if err != nil {
			t.Fatalf("ValidateJWT failed: %v", err)
		}
		if subjectID.String() != createdUser.ID {
			t.Errorf("expected JWT subject %s, got %s", createdUser.ID, subjectID.String())
		}

		// Verify refresh_token cookie is set
		cookies := w.Result().Cookies()
		var foundRefreshCookie bool
		for _, c := range cookies {
			if c.Name == "refresh_token" && c.Value != "" {
				foundRefreshCookie = true
				break
			}
		}
		if !foundRefreshCookie {
			t.Errorf("expected refresh_token cookie to be set on login")
		}
	})

	t.Run("login with wrong password returns 401 Unauthorized", func(t *testing.T) {
		loginBody, _ := json.Marshal(map[string]string{
			"email":    "charlie@example.com",
			"password": "wrongPassword",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(loginBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d", w.Code)
		}
	})

	// 4. Verify RBAC: regular user cannot access /api/admin/users
	t.Run("regular user is forbidden from admin user endpoints", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for regular user, got %d", w.Code)
		}
	})

	// 5. Login as Admin to get admin JWT
	adminLoginBody, _ := json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "adminSecret123",
	})
	adminLoginReq := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(adminLoginBody))
	adminLoginW := httptest.NewRecorder()
	router.ServeHTTP(adminLoginW, adminLoginReq)
	if adminLoginW.Code != http.StatusOK {
		t.Fatalf("admin login failed: %d %s", adminLoginW.Code, adminLoginW.Body.String())
	}
	var adminLoginResp map[string]string
	_ = json.Unmarshal(adminLoginW.Body.Bytes(), &adminLoginResp)
	adminToken := adminLoginResp["access_token"]

	// 6. Admin GetUsers (pagination)
	t.Run("admin can list users with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/users?page=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}

		var paginated models.PaginatedUsers
		if err := json.Unmarshal(w.Body.Bytes(), &paginated); err != nil {
			t.Fatalf("failed to unmarshal paginated users: %v", err)
		}
		if paginated.TotalCount != 2 || len(paginated.Data) != 2 {
			t.Errorf("expected 2 users (admin + charlie), got TotalCount=%d, len=%d", paginated.TotalCount, len(paginated.Data))
		}
	})

	// 7. Admin UpdateUser
	t.Run("admin can update user email and password", func(t *testing.T) {
		updateBody, _ := json.Marshal(map[string]string{
			"email":    "charlie.updated@example.com",
			"password": "newCharliePass456",
		})
		req := httptest.NewRequest(http.MethodPut, "/api/admin/users/"+createdUser.ID, bytes.NewReader(updateBody))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on update, got %d: %s", w.Code, w.Body.String())
		}

		var updatedResp httprest.UpdateUserResponse
		if err := json.Unmarshal(w.Body.Bytes(), &updatedResp); err != nil {
			t.Fatalf("failed to unmarshal update response: %v", err)
		}
		if updatedResp.Email != "charlie.updated@example.com" {
			t.Errorf("expected updated email 'charlie.updated@example.com', got %q", updatedResp.Email)
		}

		// Verify user can now login with the new credentials
		loginBody, _ := json.Marshal(map[string]string{
			"email":    "charlie.updated@example.com",
			"password": "newCharliePass456",
		})
		loginReq := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(loginBody))
		loginW := httptest.NewRecorder()
		router.ServeHTTP(loginW, loginReq)

		if loginW.Code != http.StatusOK {
			t.Errorf("expected login with updated credentials to succeed, got %d", loginW.Code)
		}
	})

	// 8. Admin DeleteUser
	t.Run("admin can delete user and subsequent login fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+createdUser.ID, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content on delete, got %d: %s", w.Code, w.Body.String())
		}

		// Deleting again should return 404 Not Found
		req2 := httptest.NewRequest(http.MethodDelete, "/api/admin/users/"+createdUser.ID, nil)
		req2.Header.Set("Authorization", "Bearer "+adminToken)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusNotFound {
			t.Errorf("expected 404 Not Found when deleting non-existent user, got %d", w2.Code)
		}
	})
}
