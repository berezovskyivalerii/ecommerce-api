package config_test

import (
	"testing"

	"github.com/berezovskyivalerii/ecommerce-api/internal/config"
)

func TestNewConfig(t *testing.T) {
	cfg := config.New("postgres://localhost/db", "admin@test.com", "secret")
	if cfg.ServerPort != ":8080" {
		t.Errorf("expected ServerPort ':8080', got %q", cfg.ServerPort)
	}
	if cfg.DBURL != "postgres://localhost/db" {
		t.Errorf("expected DBURL 'postgres://localhost/db', got %q", cfg.DBURL)
	}
	if cfg.AdminEmail != "admin@test.com" {
		t.Errorf("expected AdminEmail 'admin@test.com', got %q", cfg.AdminEmail)
	}
	if cfg.AdminPassword != "secret" {
		t.Errorf("expected AdminPassword 'secret', got %q", cfg.AdminPassword)
	}
}
