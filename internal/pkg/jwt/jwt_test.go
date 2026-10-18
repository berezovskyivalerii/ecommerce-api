package jwtmanager_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"testing"
	"time"

	jwtmanager "github.com/berezovskyivalerii/ecommerce-api/internal/pkg/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestJWTManager_MakeAndValidateJWT(t *testing.T) {
	m := jwtmanager.New()
	userID := uuid.New()
	secret := "test-secret"

	// 1. Valid token creation and validation
	tokenStr, err := m.MakeJWT(userID, "admin", secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}
	if tokenStr == "" {
		t.Fatalf("expected non-empty token string")
	}

	parsedID, err := m.ValidateJWT(tokenStr, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}
	if parsedID != userID {
		t.Errorf("expected userID %v, got %v", userID, parsedID)
	}

	// 2. Wrong secret
	if _, err := m.ValidateJWT(tokenStr, "wrong-secret"); err == nil {
		t.Errorf("expected error when validating with wrong secret")
	}

	// 3. Expired token
	expiredToken, err := m.MakeJWT(userID, "user", secret, -time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}
	if _, err := m.ValidateJWT(expiredToken, secret); err == nil {
		t.Errorf("expected error for expired token")
	}

	// 4. Unexpected signing method (RS256 instead of HMAC)
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	rsaTok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	rsaTokStr, err := rsaTok.SignedString(privKey)
	if err != nil {
		t.Fatalf("failed to sign RSA token: %v", err)
	}
	if _, err := m.ValidateJWT(rsaTokStr, secret); err == nil {
		t.Errorf("expected error for unexpected signing method RS256")
	}

	// 5. Invalid UUID in Subject claim
	badSubTok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "not-a-valid-uuid",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})
	badSubStr, _ := badSubTok.SignedString([]byte(secret))
	if _, err := m.ValidateJWT(badSubStr, secret); err == nil {
		t.Errorf("expected error for invalid UUID in subject")
	}
}

func TestJWTManager_GetBearerToken(t *testing.T) {
	m := jwtmanager.New()

	// 1. Missing header
	headers := http.Header{}
	if _, err := m.GetBearerToken(headers); err == nil {
		t.Errorf("expected error when Authorization header is missing")
	}

	// 2. Invalid prefix
	headers.Set("Authorization", "Basic abc123")
	if _, err := m.GetBearerToken(headers); err == nil {
		t.Errorf("expected error when Authorization header has wrong prefix")
	}

	// 3. Valid Bearer token
	headers.Set("Authorization", "Bearer my-jwt-token")
	tok, err := m.GetBearerToken(headers)
	if err != nil {
		t.Fatalf("GetBearerToken failed: %v", err)
	}
	if tok != "my-jwt-token" {
		t.Errorf("expected 'my-jwt-token', got %q", tok)
	}
}

func TestJWTManager_MakeRefreshToken(t *testing.T) {
	m := jwtmanager.New()
	t1 := m.MakeRefreshToken()
	t2 := m.MakeRefreshToken()

	if len(t1) != 64 {
		t.Errorf("expected 64-char hex token, got len=%d (%q)", len(t1), t1)
	}
	if t1 == t2 {
		t.Errorf("expected unique random refresh tokens, got identical strings")
	}
}
