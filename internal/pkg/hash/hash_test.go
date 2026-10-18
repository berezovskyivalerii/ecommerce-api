package hasher_test

import (
	"strings"
	"testing"

	hasher "github.com/berezovskyivalerii/ecommerce-api/internal/pkg/hash"
)

func TestHasher_HashAndCheckPassword(t *testing.T) {
	h := hasher.New()

	hash, err := h.HashPassword("mySecurePassword123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("expected argon2id hash prefix, got %q", hash)
	}

	// Valid password check
	if err := h.CheckPasswordHash("mySecurePassword123", hash); err != nil {
		t.Errorf("CheckPasswordHash failed on valid password: %v", err)
	}

	// Wrong password check
	if err := h.CheckPasswordHash("wrongPassword", hash); err == nil {
		t.Errorf("expected error for wrong password, got nil")
	}

	// Malformed hash check
	if err := h.CheckPasswordHash("mySecurePassword123", "invalid-hash-format"); err == nil {
		t.Errorf("expected error for malformed hash, got nil")
	}
}
