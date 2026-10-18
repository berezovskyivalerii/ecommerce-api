package app_test

import (
	"database/sql"
	"database/sql/driver"
	"sync"
	"testing"

	"github.com/berezovskyivalerii/ecommerce-api/internal/app"
	"github.com/berezovskyivalerii/ecommerce-api/internal/config"
)

var regAppFakeOnce sync.Once

type appFakeDriver struct{}

func (d *appFakeDriver) Open(string) (driver.Conn, error)            { return &appFakeConn{}, nil }
type appFakeConn struct{}
func (c *appFakeConn) Prepare(string) (driver.Stmt, error)           { return nil, nil }
func (c *appFakeConn) Close() error                                  { return nil }
func (c *appFakeConn) Begin() (driver.Tx, error)                     { return nil, nil }

func TestApp_NewWithDB_RunAndClose(t *testing.T) {
	regAppFakeOnce.Do(func() {
		sql.Register("fakedb_app", &appFakeDriver{})
	})
	db, err := sql.Open("fakedb_app", "")
	if err != nil {
		t.Fatalf("failed to open fake db: %v", err)
	}

	cfg := &config.Config{
		ServerPort: "invalid-port-address",
	}

	application := app.NewWithDB(cfg, db)
	if application == nil {
		t.Fatalf("expected non-nil App")
	}

	// Run with invalid port returns an error immediately
	if err := application.Run(); err == nil {
		t.Errorf("expected error when running server on invalid port")
	}

	if r := app.SetupRouter(nil, nil, nil, nil); r == nil {
		t.Errorf("expected non-nil router from SetupRouter")
	}

	application.Close()
}
