package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	HashedPassword string
	Email          string
	Role           string
}

type PaginatedUsers struct {
	Data       []User `json:"user"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalPages int    `json:"total_pages"`
}
