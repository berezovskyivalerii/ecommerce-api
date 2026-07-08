package jwtmanager

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct{}

func New() *JWTManager {
	return &JWTManager{}
}

func (m *JWTManager) MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// Create claims
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Issuer:    "eccommerce-api-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		Subject:   userID.String(),
	}
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create JWT
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}

	return tokenString, nil
}

func (m *JWTManager) ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error parsing token: %v", err)
	}

	if !token.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error parsing uuid in JWT subject: %v", err)
	}

	return id, nil
}

func (m *JWTManager) GetBearerToken(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	if header == "" {
		return "", fmt.Errorf("failed to get 'Authorization' header")
	}
	cuttedHeader, found := strings.CutPrefix(header, "Bearer ")
	if !found {
		return "", fmt.Errorf("invalid header format: expected: 'Bearer TOKEN_STRING', got: %s", header)
	}
	return cuttedHeader, nil
}

func (m *JWTManager) MakeRefreshToken() string {
	data := make([]byte, 32)
	rand.Read(data)
	return hex.EncodeToString(data)
}
