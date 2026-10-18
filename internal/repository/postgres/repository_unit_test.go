package postgres_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
	"github.com/google/uuid"
)

// fakeDriver implements database/sql/driver interfaces to test all sqlc generated Queries, SQLStore, ExecTx, and Repository wrappers without an external DB daemon.
var (
	registerFakeDriverOnce sync.Once
	scriptMu               sync.Mutex
	currentScript          = &fakeScript{}
)

type fakeScript struct {
	cols      []string
	rows      [][]driver.Value
	queryErr  error
	execErr   error
	beginErr  error
	commitErr error
	rbErr     error
	closeErr  error
	nextErr   error
}

func setScript(s fakeScript) {
	scriptMu.Lock()
	defer scriptMu.Unlock()
	*currentScript = s
}

type fakeDriver struct{}

func (d *fakeDriver) Open(name string) (driver.Conn, error) {
	return &fakeConn{}, nil
}

type fakeConn struct{}

func (c *fakeConn) Prepare(query string) (driver.Stmt, error) {
	return &fakeStmt{query: query}, nil
}
func (c *fakeConn) Close() error { return nil }
func (c *fakeConn) Begin() (driver.Tx, error) {
	scriptMu.Lock()
	err := currentScript.beginErr
	scriptMu.Unlock()
	if err != nil {
		return nil, err
	}
	return &fakeTx{}, nil
}

type fakeTx struct{}

func (t *fakeTx) Commit() error {
	scriptMu.Lock()
	defer scriptMu.Unlock()
	return currentScript.commitErr
}
func (t *fakeTx) Rollback() error {
	scriptMu.Lock()
	defer scriptMu.Unlock()
	return currentScript.rbErr
}

type fakeStmt struct {
	query string
}

func (s *fakeStmt) Close() error  { return nil }
func (s *fakeStmt) NumInput() int { return -1 }

func (s *fakeStmt) Exec(args []driver.Value) (driver.Result, error) {
	scriptMu.Lock()
	defer scriptMu.Unlock()
	if currentScript.execErr != nil {
		return nil, currentScript.execErr
	}
	return driver.RowsAffected(1), nil
}

func (s *fakeStmt) Query(args []driver.Value) (driver.Rows, error) {
	scriptMu.Lock()
	defer scriptMu.Unlock()
	if currentScript.queryErr != nil {
		return nil, currentScript.queryErr
	}
	rowsCopy := make([][]driver.Value, len(currentScript.rows))
	copy(rowsCopy, currentScript.rows)
	return &fakeRows{
		cols:     currentScript.cols,
		rows:     rowsCopy,
		closeErr: currentScript.closeErr,
		nextErr:  currentScript.nextErr,
	}, nil
}

type fakeRows struct {
	cols     []string
	rows     [][]driver.Value
	idx      int
	closeErr error
	nextErr  error
}

func (r *fakeRows) Columns() []string { return r.cols }
func (r *fakeRows) Close() error      { return r.closeErr }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.nextErr != nil {
		return r.nextErr
	}
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	for i, v := range r.rows[r.idx] {
		dest[i] = v
	}
	r.idx++
	return nil
}

func openFakeDB(t *testing.T) *sql.DB {
	t.Helper()
	registerFakeDriverOnce.Do(func() {
		sql.Register("fakedb_repo", &fakeDriver{})
	})
	db, err := sql.Open("fakedb_repo", "test")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestModels_EnumsScanAndValue(t *testing.T) {
	// OrderStatus
	var os postgres.OrderStatus
	if err := os.Scan([]byte("paid")); err != nil || os != postgres.OrderStatusPaid {
		t.Errorf("OrderStatus Scan([]byte) failed: %v, %v", os, err)
	}
	if err := os.Scan("shipped"); err != nil || os != postgres.OrderStatusShipped {
		t.Errorf("OrderStatus Scan(string) failed: %v, %v", os, err)
	}
	if err := os.Scan(123); err == nil {
		t.Errorf("expected error for unsupported OrderStatus Scan type")
	}

	var nos postgres.NullOrderStatus
	if err := nos.Scan(nil); err != nil || nos.Valid {
		t.Errorf("NullOrderStatus Scan(nil) failed")
	}
	if val, err := nos.Value(); err != nil || val != nil {
		t.Errorf("NullOrderStatus Value() for invalid expected nil")
	}
	if err := nos.Scan("delivered"); err != nil || !nos.Valid || nos.OrderStatus != postgres.OrderStatusDelivered {
		t.Errorf("NullOrderStatus Scan(string) failed")
	}
	if val, err := nos.Value(); err != nil || val != "delivered" {
		t.Errorf("NullOrderStatus Value() failed: %v, %v", val, err)
	}

	// PaymentStatus
	var ps postgres.PaymentStatus
	if err := ps.Scan([]byte("pending")); err != nil || ps != postgres.PaymentStatusPending {
		t.Errorf("PaymentStatus Scan([]byte) failed")
	}
	if err := ps.Scan("succeeded"); err != nil || ps != postgres.PaymentStatusSucceeded {
		t.Errorf("PaymentStatus Scan(string) failed")
	}
	if err := ps.Scan(123); err == nil {
		t.Errorf("expected error for unsupported PaymentStatus Scan type")
	}

	var nps postgres.NullPaymentStatus
	if err := nps.Scan(nil); err != nil || nps.Valid {
		t.Errorf("NullPaymentStatus Scan(nil) failed")
	}
	if val, err := nps.Value(); err != nil || val != nil {
		t.Errorf("NullPaymentStatus Value() for invalid expected nil")
	}
	if err := nps.Scan("failed"); err != nil || !nps.Valid || nps.PaymentStatus != postgres.PaymentStatusFailed {
		t.Errorf("NullPaymentStatus Scan(string) failed")
	}
	if val, err := nps.Value(); err != nil || val != "failed" {
		t.Errorf("NullPaymentStatus Value() failed: %v, %v", val, err)
	}

	// UserRole
	var ur postgres.UserRole
	if err := ur.Scan([]byte("user")); err != nil || ur != postgres.UserRoleUser {
		t.Errorf("UserRole Scan([]byte) failed")
	}
	if err := ur.Scan("admin"); err != nil || ur != postgres.UserRoleAdmin {
		t.Errorf("UserRole Scan(string) failed")
	}
	if err := ur.Scan(123); err == nil {
		t.Errorf("expected error for unsupported UserRole Scan type")
	}

	var nur postgres.NullUserRole
	if err := nur.Scan(nil); err != nil || nur.Valid {
		t.Errorf("NullUserRole Scan(nil) failed")
	}
	if val, err := nur.Value(); err != nil || val != nil {
		t.Errorf("NullUserRole Value() for invalid expected nil")
	}
	if err := nur.Scan("admin"); err != nil || !nur.Valid || nur.UserRole != postgres.UserRoleAdmin {
		t.Errorf("NullUserRole Scan(string) failed")
	}
	if val, err := nur.Value(); err != nil || val != "admin" {
		t.Errorf("NullUserRole Value() failed: %v, %v", val, err)
	}
}

func TestRepositoriesAndStore_AllMethods(t *testing.T) {
	db := openFakeDB(t)
	ctx := context.Background()
	store := postgres.NewStore(db)
	now := time.Now().UTC()
	uid := uuid.New()

	// Test Queries.WithTx
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("db.Begin failed: %v", err)
	}
	q := postgres.New(db).WithTx(tx)
	if q == nil {
		t.Fatalf("expected non-nil Queries from WithTx")
	}
	_ = tx.Commit()

	// --- Store.ExecTx ---
	setScript(fakeScript{beginErr: errors.New("begin err")})
	if err := store.ExecTx(ctx, func(postgres.Querier) error { return nil }); err == nil {
		t.Errorf("expected error when BeginTx fails")
	}

	setScript(fakeScript{})
	if err := store.ExecTx(ctx, func(postgres.Querier) error { return errors.New("fn err") }); err == nil {
		t.Errorf("expected error when tx callback fails")
	}

	setScript(fakeScript{rbErr: errors.New("rb err")})
	if err := store.ExecTx(ctx, func(postgres.Querier) error { return errors.New("fn err") }); err == nil || !strings.Contains(err.Error(), "rb err") {
		t.Errorf("expected combined tx and rb error, got %v", err)
	}

	// --- UserRepository ---
	uRepo := postgres.NewUserRepository(store)

	setScript(fakeScript{
		cols: []string{"id", "created_at", "updated_at", "email", "role"},
		rows: [][]driver.Value{{uid.String(), now, now, "u@test.com", "user"}},
	})
	if _, err := uRepo.CreateUser(ctx, postgres.CreateUserParams{Email: "u@test.com", HashedPassword: "h"}); err != nil {
		t.Errorf("CreateUser failed: %v", err)
	}
	if err := uRepo.ExampleTx(ctx, postgres.CreateUserParams{Email: "u@test.com", HashedPassword: "h"}); err != nil {
		t.Errorf("ExampleTx failed: %v", err)
	}

	setScript(fakeScript{queryErr: errors.New("q err")})
	if err := uRepo.ExampleTx(ctx, postgres.CreateUserParams{}); err == nil {
		t.Errorf("expected ExampleTx error when CreateUser fails")
	}

	setScript(fakeScript{
		cols: []string{"id", "created_at", "updated_at", "email", "role"},
		rows: [][]driver.Value{{uid.String(), now, now, "admin@test.com", "admin"}},
	})
	if _, err := uRepo.CreateAdmin(ctx, postgres.CreateAdminParams{Email: "admin@test.com", HashedPassword: "h"}); err != nil {
		t.Errorf("CreateAdmin failed: %v", err)
	}

	setScript(fakeScript{
		cols: []string{"id", "created_at", "updated_at", "email", "role"},
		rows: [][]driver.Value{{uid.String(), now, now, "u2@test.com", "user"}},
	})
	if _, err := uRepo.UpdateUser(ctx, postgres.UpdateUserParams{ID: uid, Email: "u2@test.com", HashedPassword: "h"}); err != nil {
		t.Errorf("UpdateUser failed: %v", err)
	}

	setScript(fakeScript{
		cols: []string{"count"},
		rows: [][]driver.Value{{int64(5)}},
	})
	if cnt, err := uRepo.CountUsers(ctx); err != nil || cnt != 5 {
		t.Errorf("CountUsers failed: %d, %v", cnt, err)
	}

	setScript(fakeScript{
		cols: []string{"id", "created_at", "updated_at", "hashed_password", "role"},
		rows: [][]driver.Value{{uid.String(), now, now, "h", "user"}},
	})
	if _, err := uRepo.GetUserByEmail(ctx, "u@test.com"); err != nil {
		t.Errorf("GetUserByEmail failed: %v", err)
	}

	setScript(fakeScript{
		cols: []string{"id", "created_at", "updated_at", "hashed_password", "email", "role"},
		rows: [][]driver.Value{{uid.String(), now, now, "h", "u@test.com", "user"}},
	})
	if _, err := uRepo.GetUserByID(ctx, uid); err != nil {
		t.Errorf("GetUserByID failed: %v", err)
	}

	setScript(fakeScript{
		cols: []string{"id", "created_at", "updated_at", "email", "role"},
		rows: [][]driver.Value{{uid.String(), now, now, "u@test.com", "user"}},
	})
	if list, err := uRepo.GetUsers(ctx, postgres.GetUsersParams{Limit: 10, Offset: 0}); err != nil || len(list) != 1 {
		t.Errorf("GetUsers failed: %v, %v", list, err)
	}
	setScript(fakeScript{queryErr: errors.New("err")})
	if _, err := uRepo.GetUsers(ctx, postgres.GetUsersParams{}); err == nil {
		t.Errorf("expected error on GetUsers query failure")
	}
	setScript(fakeScript{
		cols: []string{"id", "created_at", "updated_at", "email", "role"},
		rows: [][]driver.Value{{"bad-uuid", now, now, "u@test.com", "user"}},
	})
	if _, err := uRepo.GetUsers(ctx, postgres.GetUsersParams{}); err == nil {
		t.Errorf("expected scan error on GetUsers")
	}

	setScript(fakeScript{})
	if err := uRepo.DeleteUser(ctx, uid); err != nil {
		t.Errorf("DeleteUser failed: %v", err)
	}

	// --- RefreshTokensRepository ---
	rtRepo := postgres.NewRefreshTokensRepository(store)
	setScript(fakeScript{
		cols: []string{"token", "created_at", "updated_at", "user_id", "expires_at", "revoked_at"},
		rows: [][]driver.Value{{"tok1", now, now, uid.String(), now, nil}},
	})
	if _, err := rtRepo.CreateRefreshToken(ctx, postgres.CreateRefreshTokenParams{Token: "tok1", UserID: uid, ExpiresAt: now}); err != nil {
		t.Errorf("CreateRefreshToken failed: %v", err)
	}
	setScript(fakeScript{
		cols: []string{"token", "created_at", "updated_at", "user_id", "expires_at", "revoked_at"},
		rows: [][]driver.Value{{"tok1", now, now, uid.String(), now, nil}},
	})
	if _, err := rtRepo.GetRefreshToken(ctx, "tok1"); err != nil {
		t.Errorf("GetRefreshToken failed: %v", err)
	}
	setScript(fakeScript{})
	if err := rtRepo.RevokeToken(ctx, "tok1"); err != nil {
		t.Errorf("RevokeToken failed: %v", err)
	}

	// --- CategoriesRepository ---
	cRepo := postgres.NewCategoryRepository(store)
	setScript(fakeScript{
		cols: []string{"id", "name", "created_at", "updated_at"},
		rows: [][]driver.Value{{int64(1), "Books", now, now}},
	})
	if _, err := cRepo.CreateCategory(ctx, "Books"); err != nil {
		t.Errorf("CreateCategory failed: %v", err)
	}
	setScript(fakeScript{
		cols: []string{"id", "name", "created_at", "updated_at"},
		rows: [][]driver.Value{{int64(1), "Books", now, now}},
	})
	if _, err := cRepo.GetCategoryByID(ctx, 1); err != nil {
		t.Errorf("GetCategoryByID failed: %v", err)
	}
	setScript(fakeScript{
		cols: []string{"id", "name", "created_at", "updated_at"},
		rows: [][]driver.Value{{int64(1), "Books2", now, now}},
	})
	if _, err := cRepo.UpdateCategory(ctx, postgres.UpdateCategoryParams{ID: 1, Name: "Books2"}); err != nil {
		t.Errorf("UpdateCategory failed: %v", err)
	}
	setScript(fakeScript{
		cols: []string{"id", "name", "created_at", "updated_at"},
		rows: [][]driver.Value{{int64(1), "Books2", now, now}},
	})
	if list, err := cRepo.GetCategories(ctx); err != nil || len(list) != 1 {
		t.Errorf("GetCategories failed: %v", err)
	}
	setScript(fakeScript{queryErr: errors.New("err")})
	if _, err := cRepo.GetCategories(ctx); err == nil {
		t.Errorf("expected GetCategories query error")
	}
	setScript(fakeScript{
		cols: []string{"id", "name", "created_at", "updated_at"},
		rows: [][]driver.Value{{"bad-id", "Books2", now, now}},
	})
	if _, err := cRepo.GetCategories(ctx); err == nil {
		t.Errorf("expected GetCategories scan error")
	}
	setScript(fakeScript{})
	if err := cRepo.DeleteCategory(ctx, 1); err != nil {
		t.Errorf("DeleteCategory failed: %v", err)
	}

	// --- ProductsRepository ---
	pRepo := postgres.NewProductsRepository(store)
	prodCols := []string{"id", "name", "created_at", "updated_at", "price_usd", "quantity", "category_id"}
	prodRow := []driver.Value{int64(1), "Laptop", now, now, "999.99", int64(5), int64(1)}

	setScript(fakeScript{cols: prodCols, rows: [][]driver.Value{prodRow}})
	if _, err := pRepo.CreateProduct(ctx, postgres.CreateProductParams{Name: "Laptop", PriceUsd: "999.99", Quantity: 5}); err != nil {
		t.Errorf("CreateProduct failed: %v", err)
	}

	setScript(fakeScript{cols: prodCols, rows: [][]driver.Value{prodRow}})
	if _, err := pRepo.GetProductByID(ctx, 1); err != nil {
		t.Errorf("GetProductByID failed: %v", err)
	}

	setScript(fakeScript{cols: prodCols, rows: [][]driver.Value{prodRow}})
	if _, err := pRepo.UpdateProduct(ctx, postgres.UpdateProductParams{ID: 1, Name: "Laptop"}); err != nil {
		t.Errorf("UpdateProduct failed: %v", err)
	}

	setScript(fakeScript{cols: []string{"count"}, rows: [][]driver.Value{{int64(3)}}})
	if cnt, err := pRepo.CountProducts(ctx); err != nil || cnt != 3 {
		t.Errorf("CountProducts failed: %v", err)
	}

	setScript(fakeScript{cols: []string{"count"}, rows: [][]driver.Value{{int64(2)}}})
	if cnt, err := pRepo.CountProductsByCategoryID(ctx, sql.NullInt32{Int32: 1, Valid: true}); err != nil || cnt != 2 {
		t.Errorf("CountProductsByCategoryID failed: %v", err)
	}

	setScript(fakeScript{cols: []string{"count"}, rows: [][]driver.Value{{int64(1)}}})
	if cnt, err := pRepo.CountSearchProducts(ctx, "laptop"); err != nil || cnt != 1 {
		t.Errorf("CountSearchProducts failed: %v", err)
	}

	setScript(fakeScript{cols: prodCols, rows: [][]driver.Value{prodRow}})
	if list, err := pRepo.GetProducts(ctx, postgres.GetProductsParams{Limit: 10}); err != nil || len(list) != 1 {
		t.Errorf("GetProducts failed: %v", err)
	}
	setScript(fakeScript{queryErr: errors.New("err")})
	if _, err := pRepo.GetProducts(ctx, postgres.GetProductsParams{}); err == nil {
		t.Errorf("expected GetProducts query error")
	}
	setScript(fakeScript{cols: prodCols, rows: [][]driver.Value{{"bad-id", "Laptop", now, now, "999.99", int64(5), int64(1)}}})
	if _, err := pRepo.GetProducts(ctx, postgres.GetProductsParams{}); err == nil {
		t.Errorf("expected GetProducts scan error")
	}

	setScript(fakeScript{cols: prodCols, rows: [][]driver.Value{prodRow}})
	if list, err := pRepo.GetProductsByCategoryID(ctx, postgres.GetProductsByCategoryIDParams{Limit: 10}); err != nil || len(list) != 1 {
		t.Errorf("GetProductsByCategoryID failed: %v", err)
	}
	setScript(fakeScript{queryErr: errors.New("err")})
	if _, err := pRepo.GetProductsByCategoryID(ctx, postgres.GetProductsByCategoryIDParams{}); err == nil {
		t.Errorf("expected GetProductsByCategoryID query error")
	}
	setScript(fakeScript{cols: prodCols, rows: [][]driver.Value{{"bad-id", "Laptop", now, now, "999.99", int64(5), int64(1)}}})
	if _, err := pRepo.GetProductsByCategoryID(ctx, postgres.GetProductsByCategoryIDParams{}); err == nil {
		t.Errorf("expected GetProductsByCategoryID scan error")
	}

	searchCols := []string{"id", "name", "created_at", "updated_at", "price_usd", "quantity", "category_id", "rank"}
	searchRow := []driver.Value{int64(1), "Laptop", now, now, "999.99", int64(5), int64(1), float64(0.9)}
	setScript(fakeScript{cols: searchCols, rows: [][]driver.Value{searchRow}})
	if list, err := pRepo.SearchProducts(ctx, postgres.SearchProductsParams{WebsearchToTsquery: "laptop", Limit: 10}); err != nil || len(list) != 1 {
		t.Errorf("SearchProducts failed: %v", err)
	}
	setScript(fakeScript{queryErr: errors.New("err")})
	if _, err := pRepo.SearchProducts(ctx, postgres.SearchProductsParams{}); err == nil {
		t.Errorf("expected SearchProducts query error")
	}
	setScript(fakeScript{cols: searchCols, rows: [][]driver.Value{{"bad-id", "Laptop", now, now, "999.99", int64(5), int64(1), float64(0.9)}}})
	if _, err := pRepo.SearchProducts(ctx, postgres.SearchProductsParams{}); err == nil {
		t.Errorf("expected SearchProducts scan error")
	}

	setScript(fakeScript{})
	if err := pRepo.DeleteProduct(ctx, 1); err != nil {
		t.Errorf("DeleteProduct failed: %v", err)
	}

	// Test rows.Err() branch across multi-row queries
	setScript(fakeScript{cols: prodCols, nextErr: fmt.Errorf("rows iteration error")})
	_, _ = uRepo.GetUsers(ctx, postgres.GetUsersParams{})
	_, _ = cRepo.GetCategories(ctx)
	_, _ = pRepo.GetProducts(ctx, postgres.GetProductsParams{})
	_, _ = pRepo.GetProductsByCategoryID(ctx, postgres.GetProductsByCategoryIDParams{})
	_, _ = pRepo.SearchProducts(ctx, postgres.SearchProductsParams{})
}
