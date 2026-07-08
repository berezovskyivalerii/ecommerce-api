package hasher

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

type Hasher struct{}

func New() *Hasher {
	return &Hasher{}
}

func (h *Hasher) HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("error creating hash: %v", err)
	}
	return hash, nil
}

func (h *Hasher) CheckPasswordHash(password, hash string) error {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return fmt.Errorf("error comparing password and hash: %v", err)
	}
	if !match {
		return fmt.Errorf("password are not matched")
	}
	return nil
}
