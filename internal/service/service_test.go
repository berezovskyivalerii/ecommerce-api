package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/berezovskyivalerii/ecommerce-api/internal/models"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
	"github.com/berezovskyivalerii/ecommerce-api/internal/service"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockUserDB struct {
	createUserFn     func(ctx context.Context, arg postgres.CreateUserParams) (postgres.CreateUserRow, error)
	createAdminFn    func(ctx context.Context, arg postgres.CreateAdminParams) (postgres.CreateAdminRow, error)
	countUsersFn     func(ctx context.Context) (int64, error)
	getUsersFn       func(ctx context.Context, arg postgres.GetUsersParams) ([]postgres.GetUsersRow, error)
	getUserByIDFn    func(ctx context.Context, id uuid.UUID) (postgres.GetUserByIDRow, error)
	getUserByEmailFn func(ctx context.Context, email string) (postgres.GetUserByEmailRow, error)
	updateUserFn     func(ctx context.Context, arg postgres.UpdateUserParams) (postgres.UpdateUserRow, error)
	deleteUserFn     func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserDB) CreateUser(ctx context.Context, arg postgres.CreateUserParams) (postgres.CreateUserRow, error) {
	return m.createUserFn(ctx, arg)
}
func (m *mockUserDB) CreateAdmin(ctx context.Context, arg postgres.CreateAdminParams) (postgres.CreateAdminRow, error) {
	return m.createAdminFn(ctx, arg)
}
func (m *mockUserDB) CountUsers(ctx context.Context) (int64, error) {
	return m.countUsersFn(ctx)
}
func (m *mockUserDB) GetUsers(ctx context.Context, arg postgres.GetUsersParams) ([]postgres.GetUsersRow, error) {
	return m.getUsersFn(ctx, arg)
}
func (m *mockUserDB) GetUserByID(ctx context.Context, id uuid.UUID) (postgres.GetUserByIDRow, error) {
	return m.getUserByIDFn(ctx, id)
}
func (m *mockUserDB) GetUserByEmail(ctx context.Context, email string) (postgres.GetUserByEmailRow, error) {
	return m.getUserByEmailFn(ctx, email)
}
func (m *mockUserDB) UpdateUser(ctx context.Context, arg postgres.UpdateUserParams) (postgres.UpdateUserRow, error) {
	return m.updateUserFn(ctx, arg)
}
func (m *mockUserDB) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return m.deleteUserFn(ctx, id)
}

type mockHashManager struct {
	hashFn  func(password string) (string, error)
	checkFn func(password, hash string) error
}

func (m *mockHashManager) HashPassword(password string) (string, error) {
	return m.hashFn(password)
}
func (m *mockHashManager) CheckPasswordHash(password, hash string) error {
	return m.checkFn(password, hash)
}

type mockJWTManager struct {
	makeJWTFn     func(userID uuid.UUID, role string, tokenSecret string, expiresIn time.Duration) (string, error)
	makeRefreshFn func() string
}

func (m *mockJWTManager) MakeJWT(userID uuid.UUID, role string, tokenSecret string, expiresIn time.Duration) (string, error) {
	return m.makeJWTFn(userID, role, tokenSecret, expiresIn)
}
func (m *mockJWTManager) MakeRefreshToken() string {
	return m.makeRefreshFn()
}

type mockRefreshTokensDB struct {
	createFn func(ctx context.Context, arg postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error)
	getFn    func(ctx context.Context, token string) (postgres.RefreshToken, error)
	revokeFn func(ctx context.Context, token string) error
}

func (m *mockRefreshTokensDB) CreateRefreshToken(ctx context.Context, arg postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error) {
	return m.createFn(ctx, arg)
}
func (m *mockRefreshTokensDB) GetRefreshToken(ctx context.Context, token string) (postgres.RefreshToken, error) {
	return m.getFn(ctx, token)
}
func (m *mockRefreshTokensDB) RevokeToken(ctx context.Context, token string) error {
	return m.revokeFn(ctx, token)
}

type mockCategoriesDB struct {
	createFn  func(ctx context.Context, name string) (postgres.Category, error)
	getAllFn  func(ctx context.Context) ([]postgres.Category, error)
	updateFn  func(ctx context.Context, arg postgres.UpdateCategoryParams) (postgres.Category, error)
	getByIDFn func(ctx context.Context, id int32) (postgres.Category, error)
	deleteFn  func(ctx context.Context, id int32) error
}

func (m *mockCategoriesDB) CreateCategory(ctx context.Context, name string) (postgres.Category, error) {
	return m.createFn(ctx, name)
}
func (m *mockCategoriesDB) GetCategories(ctx context.Context) ([]postgres.Category, error) {
	return m.getAllFn(ctx)
}
func (m *mockCategoriesDB) UpdateCategory(ctx context.Context, arg postgres.UpdateCategoryParams) (postgres.Category, error) {
	return m.updateFn(ctx, arg)
}
func (m *mockCategoriesDB) GetCategoryByID(ctx context.Context, id int32) (postgres.Category, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockCategoriesDB) DeleteCategory(ctx context.Context, id int32) error {
	return m.deleteFn(ctx, id)
}

type mockProductsDB struct {
	createFn          func(ctx context.Context, arg postgres.CreateProductParams) (postgres.CreateProductRow, error)
	getByIDFn         func(ctx context.Context, id int32) (postgres.GetProductByIDRow, error)
	getAllFn          func(ctx context.Context, arg postgres.GetProductsParams) ([]postgres.GetProductsRow, error)
	countFn           func(ctx context.Context) (int64, error)
	deleteFn          func(ctx context.Context, id int32) error
	updateFn          func(ctx context.Context, arg postgres.UpdateProductParams) (postgres.UpdateProductRow, error)
	getByCategoryFn   func(ctx context.Context, arg postgres.GetProductsByCategoryIDParams) ([]postgres.GetProductsByCategoryIDRow, error)
	countByCategoryFn func(ctx context.Context, categoryID sql.NullInt32) (int64, error)
	searchFn          func(ctx context.Context, arg postgres.SearchProductsParams) ([]postgres.SearchProductsRow, error)
	countSearchFn     func(ctx context.Context, websearchToTsquery string) (int64, error)
}

func (m *mockProductsDB) CreateProduct(ctx context.Context, arg postgres.CreateProductParams) (postgres.CreateProductRow, error) {
	return m.createFn(ctx, arg)
}
func (m *mockProductsDB) GetProductByID(ctx context.Context, id int32) (postgres.GetProductByIDRow, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockProductsDB) GetProducts(ctx context.Context, arg postgres.GetProductsParams) ([]postgres.GetProductsRow, error) {
	return m.getAllFn(ctx, arg)
}
func (m *mockProductsDB) CountProducts(ctx context.Context) (int64, error) {
	return m.countFn(ctx)
}
func (m *mockProductsDB) DeleteProduct(ctx context.Context, id int32) error {
	return m.deleteFn(ctx, id)
}
func (m *mockProductsDB) UpdateProduct(ctx context.Context, arg postgres.UpdateProductParams) (postgres.UpdateProductRow, error) {
	return m.updateFn(ctx, arg)
}
func (m *mockProductsDB) GetProductsByCategoryID(ctx context.Context, arg postgres.GetProductsByCategoryIDParams) ([]postgres.GetProductsByCategoryIDRow, error) {
	return m.getByCategoryFn(ctx, arg)
}
func (m *mockProductsDB) CountProductsByCategoryID(ctx context.Context, categoryID sql.NullInt32) (int64, error) {
	return m.countByCategoryFn(ctx, categoryID)
}
func (m *mockProductsDB) SearchProducts(ctx context.Context, arg postgres.SearchProductsParams) ([]postgres.SearchProductsRow, error) {
	return m.searchFn(ctx, arg)
}
func (m *mockProductsDB) CountSearchProducts(ctx context.Context, websearchToTsquery string) (int64, error) {
	return m.countSearchFn(ctx, websearchToTsquery)
}

// --- UserService Tests ---

func TestUserService_Unit(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()
	now := time.Now().UTC()

	t.Run("Register success and error branches", func(t *testing.T) {
		db := &mockUserDB{}
		hm := &mockHashManager{}
		svc := service.NewUserService(db, hm, &mockJWTManager{}, &mockRefreshTokensDB{})

		// Empty inputs
		if _, err := svc.Register(ctx, models.User{Email: ""}, "pass"); err == nil {
			t.Errorf("expected error for empty email")
		}

		// Hash error
		hm.hashFn = func(string) (string, error) { return "", errors.New("hash err") }
		if _, err := svc.Register(ctx, models.User{Email: "a@b.com"}, "pass"); err == nil {
			t.Errorf("expected error when hashing fails")
		}

		// DB error
		hm.hashFn = func(string) (string, error) { return "hashed", nil }
		db.createUserFn = func(context.Context, postgres.CreateUserParams) (postgres.CreateUserRow, error) {
			return postgres.CreateUserRow{}, errors.New("db err")
		}
		if _, err := svc.Register(ctx, models.User{Email: "a@b.com"}, "pass"); err == nil {
			t.Errorf("expected error when db fails")
		}

		// Success
		db.createUserFn = func(_ context.Context, arg postgres.CreateUserParams) (postgres.CreateUserRow, error) {
			return postgres.CreateUserRow{
				ID: uid, CreatedAt: now, UpdatedAt: now, Email: arg.Email, Role: postgres.UserRoleUser,
			}, nil
		}
		u, err := svc.Register(ctx, models.User{Email: "a@b.com"}, "pass")
		if err != nil || u.ID != uid || u.Role != "user" {
			t.Errorf("unexpected Register result: %+v, err=%v", u, err)
		}
	})

	t.Run("Login success and error branches", func(t *testing.T) {
		db := &mockUserDB{}
		hm := &mockHashManager{}
		jm := &mockJWTManager{}
		rt := &mockRefreshTokensDB{}
		svc := service.NewUserService(db, hm, jm, rt)

		if _, err := svc.Login(ctx, "", "pass"); err == nil {
			t.Errorf("expected error for empty email")
		}

		db.getUserByEmailFn = func(context.Context, string) (postgres.GetUserByEmailRow, error) {
			return postgres.GetUserByEmailRow{}, errors.New("not found")
		}
		if _, err := svc.Login(ctx, "a@b.com", "pass"); err == nil {
			t.Errorf("expected error when user not found")
		}

		db.getUserByEmailFn = func(context.Context, string) (postgres.GetUserByEmailRow, error) {
			return postgres.GetUserByEmailRow{ID: uid, HashedPassword: "h", Role: postgres.UserRoleUser}, nil
		}
		hm.checkFn = func(string, string) error { return errors.New("mismatch") }
		if _, err := svc.Login(ctx, "a@b.com", "pass"); err == nil {
			t.Errorf("expected error on password mismatch")
		}

		hm.checkFn = func(string, string) error { return nil }
		jm.makeJWTFn = func(uuid.UUID, string, string, time.Duration) (string, error) {
			return "", errors.New("jwt err")
		}
		if _, err := svc.Login(ctx, "a@b.com", "pass"); err == nil {
			t.Errorf("expected error when MakeJWT fails")
		}

		jm.makeJWTFn = func(uuid.UUID, string, string, time.Duration) (string, error) { return "acc-tok", nil }
		jm.makeRefreshFn = func() string { return "ref-tok" }
		rt.createFn = func(context.Context, postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error) {
			return postgres.RefreshToken{}, errors.New("save err")
		}
		if _, err := svc.Login(ctx, "a@b.com", "pass"); err == nil {
			t.Errorf("expected error when CreateRefreshToken fails")
		}

		rt.createFn = func(context.Context, postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error) {
			return postgres.RefreshToken{Token: "ref-tok"}, nil
		}
		pair, err := svc.Login(ctx, "a@b.com", "pass")
		if err != nil || pair.AccessToken != "acc-tok" || pair.RefreshToken != "ref-tok" {
			t.Errorf("unexpected Login result: %+v, err=%v", pair, err)
		}
	})

	t.Run("GetUsers pagination and error branches", func(t *testing.T) {
		db := &mockUserDB{}
		svc := service.NewUserService(db, &mockHashManager{}, &mockJWTManager{}, &mockRefreshTokensDB{})

		db.countUsersFn = func(context.Context) (int64, error) { return 0, errors.New("count err") }
		if _, err := svc.GetUsers(ctx, 0, 0); err == nil {
			t.Errorf("expected error when CountUsers fails")
		}

		db.countUsersFn = func(context.Context) (int64, error) { return 11, nil }
		db.getUsersFn = func(context.Context, postgres.GetUsersParams) ([]postgres.GetUsersRow, error) {
			return nil, errors.New("get err")
		}
		if _, err := svc.GetUsers(ctx, 1, 10); err == nil {
			t.Errorf("expected error when GetUsers fails")
		}

		db.getUsersFn = func(context.Context, postgres.GetUsersParams) ([]postgres.GetUsersRow, error) {
			return []postgres.GetUsersRow{
				{ID: uid, Email: "u@example.com", Role: postgres.UserRoleUser},
			}, nil
		}
		res, err := svc.GetUsers(ctx, 0, 0)
		if err != nil || res.Page != 1 || res.Limit != 10 || res.TotalPages != 2 || len(res.Data) != 1 {
			t.Errorf("unexpected GetUsers result: %+v, err=%v", res, err)
		}
	})

	t.Run("GetUserByID, UpdateUser, DeleteUser", func(t *testing.T) {
		db := &mockUserDB{}
		hm := &mockHashManager{}
		svc := service.NewUserService(db, hm, &mockJWTManager{}, &mockRefreshTokensDB{})

		if _, err := svc.GetUserByID(ctx, uuid.Nil); err == nil {
			t.Errorf("expected error for nil UUID")
		}
		db.getUserByIDFn = func(context.Context, uuid.UUID) (postgres.GetUserByIDRow, error) {
			return postgres.GetUserByIDRow{}, errors.New("not found")
		}
		if _, err := svc.GetUserByID(ctx, uid); err == nil {
			t.Errorf("expected error when GetUserByID fails")
		}
		db.getUserByIDFn = func(context.Context, uuid.UUID) (postgres.GetUserByIDRow, error) {
			return postgres.GetUserByIDRow{ID: uid, Email: "u@example.com", Role: postgres.UserRoleUser}, nil
		}
		u, err := svc.GetUserByID(ctx, uid)
		if err != nil || u.ID != uid {
			t.Errorf("unexpected GetUserByID result: %+v, err=%v", u, err)
		}

		// UpdateUser
		if _, err := svc.UpdateUser(ctx, "", "pass", uid); err == nil {
			t.Errorf("expected error for empty email")
		}
		hm.hashFn = func(string) (string, error) { return "", errors.New("hash err") }
		if _, err := svc.UpdateUser(ctx, "a@b.com", "pass", uid); err == nil {
			t.Errorf("expected error when hash fails")
		}
		hm.hashFn = func(string) (string, error) { return "hashed", nil }
		db.updateUserFn = func(context.Context, postgres.UpdateUserParams) (postgres.UpdateUserRow, error) {
			return postgres.UpdateUserRow{}, errors.New("update err")
		}
		if _, err := svc.UpdateUser(ctx, "a@b.com", "pass", uid); err == nil {
			t.Errorf("expected error when update fails")
		}
		db.updateUserFn = func(context.Context, postgres.UpdateUserParams) (postgres.UpdateUserRow, error) {
			return postgres.UpdateUserRow{ID: uid, Email: "a@b.com", Role: postgres.UserRoleUser}, nil
		}
		updated, err := svc.UpdateUser(ctx, "a@b.com", "pass", uid)
		if err != nil || updated.Email != "a@b.com" {
			t.Errorf("unexpected UpdateUser result: %+v, err=%v", updated, err)
		}

		// DeleteUser
		db.deleteUserFn = func(context.Context, uuid.UUID) error { return errors.New("del err") }
		if err := svc.DeleteUser(ctx, uid); err == nil {
			t.Errorf("expected error when delete fails")
		}
		db.deleteUserFn = func(context.Context, uuid.UUID) error { return nil }
		if err := svc.DeleteUser(ctx, uid); err != nil {
			t.Errorf("unexpected error on DeleteUser: %v", err)
		}
	})

	t.Run("SeedAdmin branches", func(t *testing.T) {
		db := &mockUserDB{}
		hm := &mockHashManager{
			hashFn: func(string) (string, error) { return "hashed", nil },
		}

		// 1. Empty credentials -> skip
		service.SeedAdmin(ctx, db, hm, "", "")

		// 2. Already exists -> skip
		db.getUserByEmailFn = func(context.Context, string) (postgres.GetUserByEmailRow, error) {
			return postgres.GetUserByEmailRow{ID: uid}, nil
		}
		service.SeedAdmin(ctx, db, hm, "admin@test.com", "pass")

		// 3. Not found -> creates admin
		var created bool
		db.getUserByEmailFn = func(context.Context, string) (postgres.GetUserByEmailRow, error) {
			return postgres.GetUserByEmailRow{}, errors.New("not found")
		}
		db.createAdminFn = func(context.Context, postgres.CreateAdminParams) (postgres.CreateAdminRow, error) {
			created = true
			return postgres.CreateAdminRow{ID: uid, Role: postgres.UserRoleAdmin}, nil
		}
		service.SeedAdmin(ctx, db, hm, "admin@test.com", "pass")
		if !created {
			t.Errorf("expected CreateAdmin to be called")
		}
	})
}

// --- RefreshTokensService Tests ---

func TestRefreshTokensService_Unit(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()
	now := time.Now().UTC()

	db := &mockRefreshTokensDB{}
	udb := &mockUserDB{}
	jm := &mockJWTManager{}
	svc := service.NewRefreshTokensService(db, udb, jm)

	t.Run("Create branches", func(t *testing.T) {
		if _, err := svc.Create(ctx, models.RefreshToken{}); err == nil {
			t.Errorf("expected error on empty token/userID")
		}

		db.createFn = func(context.Context, postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error) {
			return postgres.RefreshToken{}, errors.New("db err")
		}
		if _, err := svc.Create(ctx, models.RefreshToken{Token: "t", UserID: uid}); err == nil {
			t.Errorf("expected error when db fails")
		}

		db.createFn = func(context.Context, postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error) {
			return postgres.RefreshToken{
				Token:     "t",
				UserID:    uid,
				ExpiresAt: now.Add(time.Hour),
				RevokedAt: sql.NullTime{Time: now, Valid: true},
			}, nil
		}
		tok, err := svc.Create(ctx, models.RefreshToken{Token: "t", UserID: uid, ExpiresAt: now.Add(time.Hour)})
		if err != nil || tok.Token != "t" || tok.RevokedAt == nil {
			t.Errorf("unexpected Create result: %+v, err=%v", tok, err)
		}
	})

	t.Run("Refresh branches", func(t *testing.T) {
		if _, err := svc.Refresh(ctx, ""); err == nil {
			t.Errorf("expected error on empty token")
		}

		// GetRefreshToken error
		db.getFn = func(context.Context, string) (postgres.RefreshToken, error) {
			return postgres.RefreshToken{}, errors.New("not found")
		}
		if _, err := svc.Refresh(ctx, "tok"); err == nil {
			t.Errorf("expected error when token not found")
		}

		// Expired token
		db.getFn = func(context.Context, string) (postgres.RefreshToken, error) {
			return postgres.RefreshToken{Token: "tok", UserID: uid, ExpiresAt: now.Add(-time.Hour)}, nil
		}
		if _, err := svc.Refresh(ctx, "tok"); err == nil {
			t.Errorf("expected error when token expired")
		}

		// Revoked token
		db.getFn = func(context.Context, string) (postgres.RefreshToken, error) {
			return postgres.RefreshToken{
				Token:     "tok",
				UserID:    uid,
				ExpiresAt: now.Add(time.Hour),
				RevokedAt: sql.NullTime{Time: now, Valid: true},
			}, nil
		}
		if _, err := svc.Refresh(ctx, "tok"); err == nil {
			t.Errorf("expected error when token revoked")
		}

		// RevokeToken error
		db.getFn = func(context.Context, string) (postgres.RefreshToken, error) {
			return postgres.RefreshToken{Token: "tok", UserID: uid, ExpiresAt: now.Add(time.Hour)}, nil
		}
		db.revokeFn = func(context.Context, string) error { return errors.New("revoke err") }
		if _, err := svc.Refresh(ctx, "tok"); err == nil {
			t.Errorf("expected error when revoke fails")
		}

		// GetUserByID error
		db.revokeFn = func(context.Context, string) error { return nil }
		udb.getUserByIDFn = func(context.Context, uuid.UUID) (postgres.GetUserByIDRow, error) {
			return postgres.GetUserByIDRow{}, errors.New("user err")
		}
		if _, err := svc.Refresh(ctx, "tok"); err == nil {
			t.Errorf("expected error when user lookup fails")
		}

		// MakeJWT error
		udb.getUserByIDFn = func(context.Context, uuid.UUID) (postgres.GetUserByIDRow, error) {
			return postgres.GetUserByIDRow{ID: uid, Role: postgres.UserRoleUser}, nil
		}
		jm.makeJWTFn = func(uuid.UUID, string, string, time.Duration) (string, error) {
			return "", errors.New("jwt err")
		}
		if _, err := svc.Refresh(ctx, "tok"); err == nil {
			t.Errorf("expected error when MakeJWT fails")
		}

		// Create new refresh token error
		jm.makeJWTFn = func(uuid.UUID, string, string, time.Duration) (string, error) { return "new-acc", nil }
		jm.makeRefreshFn = func() string { return "new-ref" }
		db.createFn = func(context.Context, postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error) {
			return postgres.RefreshToken{}, errors.New("create err")
		}
		if _, err := svc.Refresh(ctx, "tok"); err == nil {
			t.Errorf("expected error when saving new refresh token fails")
		}

		// Success
		db.createFn = func(context.Context, postgres.CreateRefreshTokenParams) (postgres.RefreshToken, error) {
			return postgres.RefreshToken{Token: "new-ref", UserID: uid}, nil
		}
		pair, err := svc.Refresh(ctx, "tok")
		if err != nil || pair.AccessToken != "new-acc" || pair.RefreshToken != "new-ref" {
			t.Errorf("unexpected Refresh result: %+v, err=%v", pair, err)
		}
	})

	t.Run("Revoke branches", func(t *testing.T) {
		if err := svc.Revoke(ctx, ""); err == nil {
			t.Errorf("expected error on empty token")
		}
		db.revokeFn = func(context.Context, string) error { return errors.New("err") }
		if err := svc.Revoke(ctx, "tok"); err == nil {
			t.Errorf("expected error when RevokeToken fails")
		}
		db.revokeFn = func(context.Context, string) error { return nil }
		if err := svc.Revoke(ctx, "tok"); err != nil {
			t.Errorf("unexpected error on Revoke: %v", err)
		}
	})
}

// --- CategoriesService Tests ---

func TestCategoriesService_Unit(t *testing.T) {
	ctx := context.Background()
	db := &mockCategoriesDB{}
	svc := service.NewCategoriesService(db)

	// CreateCategory
	if _, err := svc.CreateCategory(ctx, ""); err == nil {
		t.Errorf("expected error on empty name")
	}
	db.createFn = func(context.Context, string) (postgres.Category, error) {
		return postgres.Category{}, errors.New("db err")
	}
	if _, err := svc.CreateCategory(ctx, "Books"); err == nil {
		t.Errorf("expected error when CreateCategory fails")
	}
	db.createFn = func(_ context.Context, name string) (postgres.Category, error) {
		return postgres.Category{ID: 1, Name: name}, nil
	}
	cat, err := svc.CreateCategory(ctx, "Books")
	if err != nil || cat.ID != 1 || cat.Name != "Books" {
		t.Errorf("unexpected CreateCategory result: %+v, err=%v", cat, err)
	}

	// GetCategories
	db.getAllFn = func(context.Context) ([]postgres.Category, error) { return nil, errors.New("db err") }
	if _, err := svc.GetCategories(ctx); err == nil {
		t.Errorf("expected error when GetCategories fails")
	}
	db.getAllFn = func(context.Context) ([]postgres.Category, error) {
		return []postgres.Category{{ID: 1, Name: "Books"}}, nil
	}
	cats, err := svc.GetCategories(ctx)
	if err != nil || len(cats) != 1 {
		t.Errorf("unexpected GetCategories result: %+v, err=%v", cats, err)
	}

	// GetCategoryByID
	if _, err := svc.GetCategoryByID(ctx, 0); err == nil {
		t.Errorf("expected error for id < 1")
	}
	db.getByIDFn = func(context.Context, int32) (postgres.Category, error) {
		return postgres.Category{}, errors.New("not found")
	}
	if _, err := svc.GetCategoryByID(ctx, 1); err == nil {
		t.Errorf("expected error when GetCategoryByID fails")
	}
	db.getByIDFn = func(context.Context, int32) (postgres.Category, error) {
		return postgres.Category{ID: 1, Name: "Books"}, nil
	}
	cByID, err := svc.GetCategoryByID(ctx, 1)
	if err != nil || cByID.ID != 1 {
		t.Errorf("unexpected GetCategoryByID result: %+v, err=%v", cByID, err)
	}

	// UpdateCategory
	if _, err := svc.UpdateCategory(ctx, "", 1); err == nil {
		t.Errorf("expected error on empty name")
	}
	if _, err := svc.UpdateCategory(ctx, "Books", 0); err == nil {
		t.Errorf("expected error on invalid id")
	}
	db.updateFn = func(context.Context, postgres.UpdateCategoryParams) (postgres.Category, error) {
		return postgres.Category{}, errors.New("db err")
	}
	if _, err := svc.UpdateCategory(ctx, "Books2", 1); err == nil {
		t.Errorf("expected error when UpdateCategory fails")
	}
	db.updateFn = func(_ context.Context, arg postgres.UpdateCategoryParams) (postgres.Category, error) {
		return postgres.Category{ID: arg.ID, Name: arg.Name}, nil
	}
	updCat, err := svc.UpdateCategory(ctx, "Books2", 1)
	if err != nil || updCat.Name != "Books2" {
		t.Errorf("unexpected UpdateCategory result: %+v, err=%v", updCat, err)
	}

	// DeleteCategory
	if err := svc.DeleteCategory(ctx, 0); err == nil {
		t.Errorf("expected error for id < 1")
	}
	db.deleteFn = func(context.Context, int32) error { return errors.New("db err") }
	if err := svc.DeleteCategory(ctx, 1); err == nil {
		t.Errorf("expected error when DeleteCategory fails")
	}
	db.deleteFn = func(context.Context, int32) error { return nil }
	if err := svc.DeleteCategory(ctx, 1); err != nil {
		t.Errorf("unexpected error on DeleteCategory: %v", err)
	}
}

// --- ProductsService Tests ---

func TestProductsService_Unit(t *testing.T) {
	ctx := context.Background()
	pdb := &mockProductsDB{}
	cdb := &mockCategoriesDB{}
	svc := service.NewProductsService(pdb, cdb)

	t.Run("CreateProduct and UpdateProduct", func(t *testing.T) {
		cdb.getByIDFn = func(context.Context, int32) (postgres.Category, error) {
			return postgres.Category{}, errors.New("no category")
		}
		if _, err := svc.CreateProduct(ctx, models.Product{CategoryID: 99}); err == nil {
			t.Errorf("expected error when category does not exist")
		}
		if _, err := svc.UpdateProduct(ctx, 1, models.Product{CategoryID: 99}); err == nil {
			t.Errorf("expected error when category does not exist on update")
		}

		cdb.getByIDFn = func(context.Context, int32) (postgres.Category, error) {
			return postgres.Category{ID: 1, Name: "Electronics"}, nil
		}
		pdb.createFn = func(context.Context, postgres.CreateProductParams) (postgres.CreateProductRow, error) {
			return postgres.CreateProductRow{}, errors.New("db err")
		}
		if _, err := svc.CreateProduct(ctx, models.Product{Name: "Laptop", PriceUSD: 999.99, Quantity: 5, CategoryID: 1}); err == nil {
			t.Errorf("expected error when CreateProduct db fails")
		}

		pdb.createFn = func(_ context.Context, arg postgres.CreateProductParams) (postgres.CreateProductRow, error) {
			return postgres.CreateProductRow{
				ID: 10, Name: arg.Name, PriceUsd: arg.PriceUsd, Quantity: arg.Quantity, CategoryID: arg.CategoryID,
			}, nil
		}
		created, err := svc.CreateProduct(ctx, models.Product{Name: "Laptop", PriceUSD: 999.99, Quantity: 5, CategoryID: 1})
		if err != nil || created.ID != 10 || created.PriceUSD != 999.99 {
			t.Errorf("unexpected CreateProduct result: %+v, err=%v", created, err)
		}

		pdb.updateFn = func(context.Context, postgres.UpdateProductParams) (postgres.UpdateProductRow, error) {
			return postgres.UpdateProductRow{}, errors.New("db err")
		}
		if _, err := svc.UpdateProduct(ctx, 10, models.Product{Name: "Laptop Pro", PriceUSD: 1299.50, Quantity: 3, CategoryID: 1}); err == nil {
			t.Errorf("expected error when UpdateProduct db fails")
		}

		pdb.updateFn = func(_ context.Context, arg postgres.UpdateProductParams) (postgres.UpdateProductRow, error) {
			return postgres.UpdateProductRow{
				ID: arg.ID, Name: arg.Name, PriceUsd: arg.PriceUsd, Quantity: arg.Quantity, CategoryID: arg.CategoryID,
			}, nil
		}
		updated, err := svc.UpdateProduct(ctx, 10, models.Product{Name: "Laptop Pro", PriceUSD: 1299.50, Quantity: 3, CategoryID: 1})
		if err != nil || updated.Name != "Laptop Pro" || updated.PriceUSD != 1299.50 {
			t.Errorf("unexpected UpdateProduct result: %+v, err=%v", updated, err)
		}
	})

	t.Run("GetProductByID and DeleteProduct", func(t *testing.T) {
		if _, err := svc.GetProductByID(ctx, 0); err == nil {
			t.Errorf("expected error on id < 1")
		}
		pdb.getByIDFn = func(context.Context, int32) (postgres.GetProductByIDRow, error) {
			return postgres.GetProductByIDRow{}, errors.New("not found")
		}
		if _, err := svc.GetProductByID(ctx, 10); err == nil {
			t.Errorf("expected error when GetProductByID fails")
		}
		pdb.getByIDFn = func(context.Context, int32) (postgres.GetProductByIDRow, error) {
			return postgres.GetProductByIDRow{
				ID: 10, Name: "Laptop", PriceUsd: "999.99", Quantity: 5, CategoryID: sql.NullInt32{Int32: 1, Valid: true},
			}, nil
		}
		p, err := svc.GetProductByID(ctx, 10)
		if err != nil || p.ID != 10 || p.PriceUSD != 999.99 {
			t.Errorf("unexpected GetProductByID result: %+v, err=%v", p, err)
		}

		if err := svc.DeleteProduct(ctx, 0); err == nil {
			t.Errorf("expected error on DeleteProduct with id < 1")
		}
		pdb.deleteFn = func(context.Context, int32) error { return errors.New("del err") }
		if err := svc.DeleteProduct(ctx, 10); err == nil {
			t.Errorf("expected error when DeleteProduct db fails")
		}
		pdb.deleteFn = func(context.Context, int32) error { return nil }
		if err := svc.DeleteProduct(ctx, 10); err != nil {
			t.Errorf("unexpected error on DeleteProduct: %v", err)
		}
	})

	t.Run("GetProducts, GetProductsByCategory, SearchProducts", func(t *testing.T) {
		// GetProducts
		pdb.countFn = func(context.Context) (int64, error) { return 0, errors.New("count err") }
		if _, err := svc.GetProducts(ctx, 0, 0); err == nil {
			t.Errorf("expected error when CountProducts fails")
		}
		pdb.countFn = func(context.Context) (int64, error) { return 15, nil }
		pdb.getAllFn = func(context.Context, postgres.GetProductsParams) ([]postgres.GetProductsRow, error) {
			return nil, errors.New("get err")
		}
		if _, err := svc.GetProducts(ctx, 1, 10); err == nil {
			t.Errorf("expected error when GetProducts fails")
		}
		pdb.getAllFn = func(context.Context, postgres.GetProductsParams) ([]postgres.GetProductsRow, error) {
			return []postgres.GetProductsRow{
				{ID: 1, Name: "Phone", PriceUsd: "500.00", Quantity: 2, CategoryID: sql.NullInt32{Int32: 1, Valid: true}},
			}, nil
		}
		list, err := svc.GetProducts(ctx, 0, 0)
		if err != nil || list.TotalPages != 2 || len(list.Data) != 1 {
			t.Errorf("unexpected GetProducts result: %+v, err=%v", list, err)
		}

		// GetProductsByCategory
		pdb.countByCategoryFn = func(context.Context, sql.NullInt32) (int64, error) { return 0, errors.New("err") }
		if _, err := svc.GetProductsByCategory(ctx, 1, 0, 0); err == nil {
			t.Errorf("expected error when CountProductsByCategoryID fails")
		}
		pdb.countByCategoryFn = func(context.Context, sql.NullInt32) (int64, error) { return 12, nil }
		pdb.getByCategoryFn = func(context.Context, postgres.GetProductsByCategoryIDParams) ([]postgres.GetProductsByCategoryIDRow, error) {
			return nil, errors.New("err")
		}
		if _, err := svc.GetProductsByCategory(ctx, 1, 1, 10); err == nil {
			t.Errorf("expected error when GetProductsByCategoryID fails")
		}
		pdb.getByCategoryFn = func(context.Context, postgres.GetProductsByCategoryIDParams) ([]postgres.GetProductsByCategoryIDRow, error) {
			return []postgres.GetProductsByCategoryIDRow{
				{ID: 1, Name: "Phone", PriceUsd: "500.00", Quantity: 2, CategoryID: sql.NullInt32{Int32: 1, Valid: true}},
			}, nil
		}
		byCat, err := svc.GetProductsByCategory(ctx, 1, 0, 0)
		if err != nil || byCat.TotalPages != 2 || len(byCat.Data) != 1 {
			t.Errorf("unexpected GetProductsByCategory result: %+v, err=%v", byCat, err)
		}

		// SearchProducts
		pdb.countSearchFn = func(context.Context, string) (int64, error) { return 0, errors.New("err") }
		if _, err := svc.SearchProducts(ctx, "phone", 0, 0); err == nil {
			t.Errorf("expected error when CountSearchProducts fails")
		}
		pdb.countSearchFn = func(context.Context, string) (int64, error) { return 11, nil }
		pdb.searchFn = func(context.Context, postgres.SearchProductsParams) ([]postgres.SearchProductsRow, error) {
			return nil, errors.New("err")
		}
		if _, err := svc.SearchProducts(ctx, "phone", 1, 10); err == nil {
			t.Errorf("expected error when SearchProducts fails")
		}
		pdb.searchFn = func(context.Context, postgres.SearchProductsParams) ([]postgres.SearchProductsRow, error) {
			return []postgres.SearchProductsRow{
				{ID: 1, Name: "Phone", PriceUsd: "500.00", Quantity: 2, CategoryID: sql.NullInt32{Int32: 1, Valid: true}, Rank: 0.9},
			}, nil
		}
		searched, err := svc.SearchProducts(ctx, "phone", 0, 0)
		if err != nil || searched.TotalPages != 2 || len(searched.Data) != 1 {
			t.Errorf("unexpected SearchProducts result: %+v, err=%v", searched, err)
		}
	})
}
