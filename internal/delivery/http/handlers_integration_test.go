package httprest_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

// fullInMemoryBackend implements UserDB, RefreshTokensDB, CategoriesDB, and ProductsDB for end-to-end HTTP integration tests.
type fullInMemoryBackend struct {
	*inMemoryUserDB
	mu         sync.RWMutex
	categories map[int32]postgres.Category
	nextCatID  int32
	products   map[int32]postgres.Product
	nextProdID int32
}

func newFullInMemoryBackend() *fullInMemoryBackend {
	return &fullInMemoryBackend{
		inMemoryUserDB: newInMemoryUserDB(),
		categories:     make(map[int32]postgres.Category),
		nextCatID:      1,
		products:       make(map[int32]postgres.Product),
		nextProdID:     1,
	}
}

// RefreshTokensDB methods on fullInMemoryBackend

func (b *fullInMemoryBackend) GetRefreshToken(_ context.Context, token string) (postgres.RefreshToken, error) {
	b.inMemoryUserDB.mu.RLock()
	defer b.inMemoryUserDB.mu.RUnlock()

	rt, ok := b.refreshTokens[token]
	if !ok {
		return postgres.RefreshToken{}, errors.New("not found")
	}
	return rt, nil
}

func (b *fullInMemoryBackend) RevokeToken(_ context.Context, token string) error {
	b.inMemoryUserDB.mu.Lock()
	defer b.inMemoryUserDB.mu.Unlock()

	rt, ok := b.refreshTokens[token]
	if !ok {
		return errors.New("not found")
	}
	now := time.Now().UTC()
	rt.RevokedAt = sql.NullTime{Time: now, Valid: true}
	rt.UpdatedAt = now
	b.refreshTokens[token] = rt
	return nil
}

// CategoriesDB methods

func (b *fullInMemoryBackend) CreateCategory(_ context.Context, name string) (postgres.Category, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, c := range b.categories {
		if c.Name == name {
			return postgres.Category{}, errors.New("duplicate category")
		}
	}
	now := time.Now().UTC()
	c := postgres.Category{
		ID:        b.nextCatID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	b.categories[c.ID] = c
	b.nextCatID++
	return c, nil
}

func (b *fullInMemoryBackend) GetCategories(_ context.Context) ([]postgres.Category, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	res := make([]postgres.Category, 0, len(b.categories))
	for _, c := range b.categories {
		res = append(res, c)
	}
	return res, nil
}

func (b *fullInMemoryBackend) GetCategoryByID(_ context.Context, id int32) (postgres.Category, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	c, ok := b.categories[id]
	if !ok {
		return postgres.Category{}, errors.New("not found")
	}
	return c, nil
}

func (b *fullInMemoryBackend) UpdateCategory(_ context.Context, arg postgres.UpdateCategoryParams) (postgres.Category, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	c, ok := b.categories[arg.ID]
	if !ok {
		return postgres.Category{}, errors.New("not found")
	}
	c.Name = arg.Name
	c.UpdatedAt = time.Now().UTC()
	b.categories[arg.ID] = c
	return c, nil
}

func (b *fullInMemoryBackend) DeleteCategory(_ context.Context, id int32) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.categories[id]; !ok {
		return errors.New("not found")
	}
	delete(b.categories, id)
	return nil
}

// ProductsDB methods

func (b *fullInMemoryBackend) CreateProduct(_ context.Context, arg postgres.CreateProductParams) (postgres.CreateProductRow, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now().UTC()
	p := postgres.Product{
		ID:         b.nextProdID,
		Name:       arg.Name,
		CreatedAt:  now,
		UpdatedAt:  now,
		PriceUsd:   arg.PriceUsd,
		Quantity:   arg.Quantity,
		CategoryID: arg.CategoryID,
	}
	b.products[p.ID] = p
	b.nextProdID++
	return postgres.CreateProductRow{
		ID:         p.ID,
		Name:       p.Name,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		PriceUsd:   p.PriceUsd,
		Quantity:   p.Quantity,
		CategoryID: p.CategoryID,
	}, nil
}

func (b *fullInMemoryBackend) GetProductByID(_ context.Context, id int32) (postgres.GetProductByIDRow, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	p, ok := b.products[id]
	if !ok {
		return postgres.GetProductByIDRow{}, errors.New("not found")
	}
	return postgres.GetProductByIDRow{
		ID:         p.ID,
		Name:       p.Name,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		PriceUsd:   p.PriceUsd,
		Quantity:   p.Quantity,
		CategoryID: p.CategoryID,
	}, nil
}

func (b *fullInMemoryBackend) GetProducts(_ context.Context, arg postgres.GetProductsParams) ([]postgres.GetProductsRow, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	all := make([]postgres.GetProductsRow, 0, len(b.products))
	for _, p := range b.products {
		all = append(all, postgres.GetProductsRow{
			ID:         p.ID,
			Name:       p.Name,
			CreatedAt:  p.CreatedAt,
			UpdatedAt:  p.UpdatedAt,
			PriceUsd:   p.PriceUsd,
			Quantity:   p.Quantity,
			CategoryID: p.CategoryID,
		})
	}
	start := int(arg.Offset)
	if start >= len(all) {
		return []postgres.GetProductsRow{}, nil
	}
	end := start + int(arg.Limit)
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], nil
}

func (b *fullInMemoryBackend) CountProducts(_ context.Context) (int64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return int64(len(b.products)), nil
}

func (b *fullInMemoryBackend) DeleteProduct(_ context.Context, id int32) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.products[id]; !ok {
		return errors.New("not found")
	}
	delete(b.products, id)
	return nil
}

func (b *fullInMemoryBackend) UpdateProduct(_ context.Context, arg postgres.UpdateProductParams) (postgres.UpdateProductRow, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	p, ok := b.products[arg.ID]
	if !ok {
		return postgres.UpdateProductRow{}, errors.New("not found")
	}
	p.Name = arg.Name
	p.PriceUsd = arg.PriceUsd
	p.Quantity = arg.Quantity
	p.CategoryID = arg.CategoryID
	p.UpdatedAt = time.Now().UTC()
	b.products[arg.ID] = p

	return postgres.UpdateProductRow{
		ID:         p.ID,
		Name:       p.Name,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		PriceUsd:   p.PriceUsd,
		Quantity:   p.Quantity,
		CategoryID: p.CategoryID,
	}, nil
}

func (b *fullInMemoryBackend) GetProductsByCategoryID(_ context.Context, arg postgres.GetProductsByCategoryIDParams) ([]postgres.GetProductsByCategoryIDRow, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	res := []postgres.GetProductsByCategoryIDRow{}
	for _, p := range b.products {
		if p.CategoryID.Int32 == arg.CategoryID.Int32 {
			res = append(res, postgres.GetProductsByCategoryIDRow{
				ID:         p.ID,
				Name:       p.Name,
				CreatedAt:  p.CreatedAt,
				UpdatedAt:  p.UpdatedAt,
				PriceUsd:   p.PriceUsd,
				Quantity:   p.Quantity,
				CategoryID: p.CategoryID,
			})
		}
	}
	return res, nil
}

func (b *fullInMemoryBackend) CountProductsByCategoryID(_ context.Context, categoryID sql.NullInt32) (int64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var cnt int64
	for _, p := range b.products {
		if p.CategoryID.Int32 == categoryID.Int32 {
			cnt++
		}
	}
	return cnt, nil
}

func (b *fullInMemoryBackend) SearchProducts(_ context.Context, arg postgres.SearchProductsParams) ([]postgres.SearchProductsRow, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	q := strings.ToLower(arg.WebsearchToTsquery)
	res := []postgres.SearchProductsRow{}
	for _, p := range b.products {
		if strings.Contains(strings.ToLower(p.Name), q) {
			res = append(res, postgres.SearchProductsRow{
				ID:         p.ID,
				Name:       p.Name,
				CreatedAt:  p.CreatedAt,
				UpdatedAt:  p.UpdatedAt,
				PriceUsd:   p.PriceUsd,
				Quantity:   p.Quantity,
				CategoryID: p.CategoryID,
				Rank:       1.0,
			})
		}
	}
	return res, nil
}

func (b *fullInMemoryBackend) CountSearchProducts(_ context.Context, websearchToTsquery string) (int64, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	q := strings.ToLower(websearchToTsquery)
	var cnt int64
	for _, p := range b.products {
		if strings.Contains(strings.ToLower(p.Name), q) {
			cnt++
		}
	}
	return cnt, nil
}

func setupFullAppRouter() (*gin.Engine, *fullInMemoryBackend, *jwtmanager.JWTManager) {
	gin.SetMode(gin.TestMode)
	backend := newFullInMemoryBackend()
	passwordHasher := hasher.New()
	jwtMgr := jwtmanager.New()

	service.SeedAdmin(context.Background(), backend, passwordHasher, "admin@shop.com", "admin123")

	userSvc := service.NewUserService(backend, passwordHasher, jwtMgr, backend)
	refreshSvc := service.NewRefreshTokensService(backend, backend, jwtMgr)
	catSvc := service.NewCategoriesService(backend)
	prodSvc := service.NewProductsService(backend, backend)

	userH := httprest.NewUserHandlers(userSvc)
	refreshH := httprest.NewRefreshTokensHandlers(refreshSvc)
	catH := httprest.NewCategoriesHandlers(catSvc)
	prodH := httprest.NewProductsHandlers(prodSvc)

	r := gin.New()
	r.POST("/api/register", userH.Register)
	r.POST("/api/login", userH.Login)
	r.POST("/api/refresh", refreshH.RefreshToken)
	r.POST("/api/revoke", refreshH.RevokeToken)

	r.GET("/api/categories", catH.GetCategories)
	r.GET("/api/categories/:category_id", catH.GetCategoryByID)
	r.GET("/api/categories/:category_id/products", prodH.GetProductsByCategory)

	r.GET("/api/products", prodH.GetProducts)
	r.GET("/api/products/search", prodH.SearchProducts)
	r.GET("/api/products/:id", prodH.GetProductByID)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware("secret-123"))
	{
		api.GET("/users/me", userH.GetCurrentUser)
		api.PUT("/users/me", userH.UpdateCurrentUser)

		adminRoutes := api.Group("/admin")
		adminRoutes.Use(middleware.RequireRole("admin"))
		{
			adminRoutes.GET("/users", userH.GetUsers)
			adminRoutes.GET("/users/:id", userH.GetUserByID)
			adminRoutes.PUT("/users/:id", userH.UpdateUser)
			adminRoutes.DELETE("/users/:id", userH.DeleteUser)

			adminRoutes.POST("/categories", catH.CreateCategory)
			adminRoutes.PUT("/categories/:id", catH.UpdateCategory)
			adminRoutes.DELETE("/categories/:id", catH.DeleteCategory)

			adminRoutes.POST("/products", prodH.CreateProduct)
			adminRoutes.PUT("/products/:id", prodH.UpdateProduct)
			adminRoutes.DELETE("/products/:id", prodH.DeleteProduct)
		}
	}

	return r, backend, jwtMgr
}

func TestFullAPI_Integration_TokensCategoriesProductsUsers(t *testing.T) {
	router, _, jwtMgr := setupFullAppRouter()

	// 1. Register & Login regular user, test /api/users/me & /api/refresh & /api/revoke
	regBody, _ := json.Marshal(map[string]string{"email": "buyer@shop.com", "password": "buyerPass123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", w.Code, w.Body.String())
	}
	var regUser httprest.CreateUserResponse
	_ = json.Unmarshal(w.Body.Bytes(), &regUser)

	// Login
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}
	var loginData map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &loginData)
	buyerAccess := loginData["access_token"]
	var refreshCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "refresh_token" {
			refreshCookie = c
		}
	}
	if refreshCookie == nil {
		t.Fatalf("missing refresh_token cookie")
	}

	// GET /api/users/me
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+buyerAccess)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/users/me failed: %d %s", w.Code, w.Body.String())
	}

	// PUT /api/users/me
	updMeBody, _ := json.Marshal(map[string]string{"email": "buyer2@shop.com", "password": "buyerPass456"})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/users/me", bytes.NewReader(updMeBody))
	req.Header.Set("Authorization", "Bearer "+buyerAccess)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT /api/users/me failed: %d %s", w.Code, w.Body.String())
	}

	// Refresh token flow: POST /api/refresh without cookie -> 401
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on refresh without cookie, got %d", w.Code)
	}

	// POST /api/refresh with valid cookie -> 200 and new cookie
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	req.AddCookie(refreshCookie)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/refresh failed: %d %s", w.Code, w.Body.String())
	}
	var newRefreshCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "refresh_token" {
			newRefreshCookie = c
		}
	}

	// Reusing old revoked refresh token -> 401
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/refresh", nil)
	req.AddCookie(refreshCookie)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 when reusing revoked refresh token, got %d", w.Code)
	}

	// Revoke token without cookie -> 401
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/revoke", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on revoke without cookie, got %d", w.Code)
	}

	// Revoke token with valid cookie -> 200
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/revoke", nil)
	req.AddCookie(newRefreshCookie)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/revoke failed: %d %s", w.Code, w.Body.String())
	}

	// 2. Admin Categories & Products CRUD + Search + GetByID
	adminToken, _ := jwtMgr.MakeJWT(uuid.New(), "admin", "secret-123", time.Hour)

	// GET /api/admin/users/:id
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/admin/users/"+regUser.ID, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/admin/users/:id failed: %d %s", w.Code, w.Body.String())
	}

	// Create Category
	catBody, _ := json.Marshal(map[string]string{"name": "Electronics"})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/categories", bytes.NewReader(catBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create category failed: %d %s", w.Code, w.Body.String())
	}
	var createdCat models.Category
	_ = json.Unmarshal(w.Body.Bytes(), &createdCat)

	// GetCategories & GetCategoryByID
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/categories failed: %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/categories/1", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/categories/1 failed: %d %s", w.Code, w.Body.String())
	}

	// UpdateCategory
	updCatBody, _ := json.Marshal(map[string]string{"name": "Consumer Electronics"})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/categories/1", bytes.NewReader(updCatBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT /api/admin/categories/1 failed: %d %s", w.Code, w.Body.String())
	}

	// Create Product
	prodBody, _ := json.Marshal(map[string]any{
		"name":        "Gaming Laptop",
		"price_usd":   1499.99,
		"quantity":    10,
		"category_id": createdCat.ID,
	})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/products", bytes.NewReader(prodBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST /api/admin/products failed: %d %s", w.Code, w.Body.String())
	}

	// GetProducts, GetProductByID, GetProductsByCategory, SearchProducts
	for _, endpoint := range []string{
		"/api/products?page=1&limit=10",
		"/api/products/1",
		"/api/categories/1/products?page=1&limit=10",
		"/api/products/search?q=laptop&page=1&limit=10",
	} {
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, endpoint, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s failed: %d %s", endpoint, w.Code, w.Body.String())
		}
	}

	// UpdateProduct
	updProdBody, _ := json.Marshal(map[string]any{
		"name":        "Gaming Laptop Pro",
		"price_usd":   1799.99,
		"quantity":    8,
		"category_id": createdCat.ID,
	})
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/admin/products/1", bytes.NewReader(updProdBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT /api/admin/products/1 failed: %d %s", w.Code, w.Body.String())
	}

	// DeleteProduct & DeleteCategory
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/admin/products/1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/admin/products/1 failed: %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/admin/categories/1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("DELETE /api/admin/categories/1 failed: %d", w.Code)
	}
}

// --- Handler Error/Validation Branch Unit Tests ---

type errUserService struct{}

func (errUserService) Register(context.Context, models.User, string) (models.User, error) {
	return models.User{}, errors.New("err")
}
func (errUserService) Login(context.Context, string, string) (models.TokenPair, error) {
	return models.TokenPair{}, errors.New("err")
}
func (errUserService) GetUsers(context.Context, int, int) (models.PaginatedUsers, error) {
	return models.PaginatedUsers{}, errors.New("err")
}
func (errUserService) GetUserByID(context.Context, uuid.UUID) (models.User, error) {
	return models.User{}, errors.New("err")
}
func (errUserService) UpdateUser(context.Context, string, string, uuid.UUID) (models.User, error) {
	return models.User{}, errors.New("err")
}
func (errUserService) DeleteUser(context.Context, uuid.UUID) error {
	return errors.New("err")
}

type errCategoriesService struct{}

func (errCategoriesService) CreateCategory(context.Context, string) (models.Category, error) {
	return models.Category{}, errors.New("err")
}
func (errCategoriesService) UpdateCategory(context.Context, string, int32) (models.Category, error) {
	return models.Category{}, errors.New("err")
}
func (errCategoriesService) GetCategories(context.Context) ([]models.Category, error) {
	return nil, errors.New("err")
}
func (errCategoriesService) GetCategoryByID(context.Context, int32) (postgres.Category, error) {
	return postgres.Category{}, errors.New("err")
}
func (errCategoriesService) DeleteCategory(context.Context, int32) error {
	return errors.New("err")
}

type errProductsService struct{}

func (errProductsService) CreateProduct(context.Context, models.Product) (models.Product, error) {
	return models.Product{}, errors.New("err")
}
func (errProductsService) GetProducts(context.Context, int, int) (models.PaginatedProducts, error) {
	return models.PaginatedProducts{}, errors.New("err")
}
func (errProductsService) GetProductByID(context.Context, int32) (models.Product, error) {
	return models.Product{}, errors.New("err")
}
func (errProductsService) DeleteProduct(context.Context, int32) error {
	return errors.New("err")
}
func (errProductsService) UpdateProduct(context.Context, int32, models.Product) (models.Product, error) {
	return models.Product{}, errors.New("err")
}
func (errProductsService) GetProductsByCategory(context.Context, int32, int, int) (models.PaginatedProducts, error) {
	return models.PaginatedProducts{}, errors.New("err")
}
func (errProductsService) SearchProducts(context.Context, string, int, int) (models.PaginatedProducts, error) {
	return models.PaginatedProducts{}, errors.New("err")
}

type errRefreshService struct{}

func (errRefreshService) Refresh(context.Context, string) (models.TokenPair, error) {
	return models.TokenPair{}, errors.New("err")
}
func (errRefreshService) Revoke(context.Context, string) error {
	return errors.New("err")
}

func TestHandlers_ErrorAndValidationBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uh := httprest.NewUserHandlers(errUserService{})
	ch := httprest.NewCategoriesHandlers(errCategoriesService{})
	ph := httprest.NewProductsHandlers(errProductsService{})
	rh := httprest.NewRefreshTokensHandlers(errRefreshService{})

	r := gin.New()
	r.POST("/login", uh.Login)
	r.GET("/users", uh.GetUsers)
	r.GET("/users/me", uh.GetCurrentUser)
	r.PUT("/users/me", uh.UpdateCurrentUser)
	r.GET("/users/:id", uh.GetUserByID)
	r.PUT("/users/:id", uh.UpdateUser)
	r.DELETE("/users/:id", uh.DeleteUser)

	r.POST("/revoke", rh.RevokeToken)

	r.GET("/categories", ch.GetCategories)
	r.GET("/categories/:category_id", ch.GetCategoryByID)
	r.POST("/categories", ch.CreateCategory)
	r.PUT("/categories/:id", ch.UpdateCategory)
	r.DELETE("/categories/:id", ch.DeleteCategory)

	r.GET("/products", ph.GetProducts)
	r.GET("/products/search", ph.SearchProducts)
	r.GET("/products/:id", ph.GetProductByID)
	r.GET("/categories/:category_id/products", ph.GetProductsByCategory)
	r.POST("/products", ph.CreateProduct)
	r.PUT("/products/:id", ph.UpdateProduct)
	r.DELETE("/products/:id", ph.DeleteProduct)

	cases := []struct {
		method string
		path   string
		body   string
		cookie *http.Cookie
		status int
	}{
		{http.MethodPost, "/login", "{bad", nil, http.StatusBadRequest},
		{http.MethodGet, "/users?page=-1&limit=999", "", nil, http.StatusInternalServerError},
		{http.MethodGet, "/users/me", "", nil, http.StatusUnauthorized},
		{http.MethodPut, "/users/me", `{}`, nil, http.StatusUnauthorized},
		{http.MethodGet, "/users/not-uuid", "", nil, http.StatusBadRequest},
		{http.MethodGet, "/users/" + uuid.NewString(), "", nil, http.StatusNotFound},
		{http.MethodPut, "/users/not-uuid", `{}`, nil, http.StatusBadRequest},
		{http.MethodPut, "/users/" + uuid.NewString(), `{bad`, nil, http.StatusBadRequest},
		{http.MethodPut, "/users/" + uuid.NewString(), `{"email":"a@b.com","password":"password123"}`, nil, http.StatusInternalServerError},
		{http.MethodDelete, "/users/not-uuid", "", nil, http.StatusBadRequest},

		{http.MethodPost, "/revoke", "", &http.Cookie{Name: "refresh_token", Value: "tok"}, http.StatusInternalServerError},

		{http.MethodGet, "/categories", "", nil, http.StatusInternalServerError},
		{http.MethodGet, "/categories/abc", "", nil, http.StatusBadRequest},
		{http.MethodGet, "/categories/1", "", nil, http.StatusNotFound},
		{http.MethodPost, "/categories", `{bad`, nil, http.StatusBadRequest},
		{http.MethodPost, "/categories", `{"name":"Cat"}`, nil, http.StatusInternalServerError},
		{http.MethodPut, "/categories/abc", `{"name":"Cat"}`, nil, http.StatusBadRequest},
		{http.MethodPut, "/categories/1", `{bad`, nil, http.StatusBadRequest},
		{http.MethodPut, "/categories/1", `{"name":"Cat"}`, nil, http.StatusInternalServerError},
		{http.MethodDelete, "/categories/abc", "", nil, http.StatusBadRequest},
		{http.MethodDelete, "/categories/1", "", nil, http.StatusInternalServerError},

		{http.MethodGet, "/products", "", nil, http.StatusInternalServerError},
		{http.MethodGet, "/products/search", "", nil, http.StatusBadRequest},
		{http.MethodGet, "/products/search?q=test", "", nil, http.StatusInternalServerError},
		{http.MethodGet, "/products/abc", "", nil, http.StatusBadRequest},
		{http.MethodGet, "/products/1", "", nil, http.StatusNotFound},
		{http.MethodGet, "/categories/abc/products", "", nil, http.StatusBadRequest},
		{http.MethodGet, "/categories/1/products", "", nil, http.StatusInternalServerError},
		{http.MethodPost, "/products", `{bad`, nil, http.StatusBadRequest},
		{http.MethodPost, "/products", `{"name":"P","price_usd":10,"quantity":1,"category_id":1}`, nil, http.StatusInternalServerError},
		{http.MethodPut, "/products/abc", `{}`, nil, http.StatusBadRequest},
		{http.MethodPut, "/products/1", `{bad`, nil, http.StatusBadRequest},
		{http.MethodPut, "/products/1", `{"name":"P","price_usd":10,"quantity":1,"category_id":1}`, nil, http.StatusInternalServerError},
		{http.MethodDelete, "/products/abc", "", nil, http.StatusBadRequest},
		{http.MethodDelete, "/products/1", "", nil, http.StatusInternalServerError},
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		if tc.cookie != nil {
			req.AddCookie(tc.cookie)
		}
		r.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Errorf("%s %s: expected %d, got %d (%s)", tc.method, tc.path, tc.status, w.Code, w.Body.String())
		}
	}

	// Test GetCurrentUser & UpdateCurrentUser context edge cases (non-string user_id, invalid UUID, bad body, service error)
	for _, ctxVal := range []any{12345, "not-a-uuid", uuid.NewString()} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users/me", nil)
		c.Set("user_id", ctxVal)
		uh.GetCurrentUser(c)

		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest(http.MethodPut, "/users/me", strings.NewReader(`{"email":"a@b.com","password":"password123"}`))
		c2.Request.Header.Set("Content-Type", "application/json")
		c2.Set("user_id", ctxVal)
		uh.UpdateCurrentUser(c2)
	}

	// UpdateCurrentUser with invalid JSON body
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest(http.MethodPut, "/users/me", strings.NewReader(`{bad`))
	c3.Request.Header.Set("Content-Type", "application/json")
	c3.Set("user_id", uuid.NewString())
	uh.UpdateCurrentUser(c3)
	if w3.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on invalid JSON in UpdateCurrentUser, got %d", w3.Code)
	}
}
