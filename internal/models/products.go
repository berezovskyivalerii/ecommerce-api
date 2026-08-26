package models

import "time"

type Product struct {
	ID         int32
	Name       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	PriceUSD   float64
	Quantity   int32
	CategotyID int32
}

type PaginatedProducts struct {
	Data       []Product `json:"data"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}
