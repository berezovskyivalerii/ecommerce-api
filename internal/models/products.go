package models

import "time"

type Product struct {
	ID         int32     `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	PriceUSD   float64   `json:"price_usd"`
	Quantity   int32     `json:"quantity"`
	CategoryID int32     `json:"category_id"`
}

type PaginatedProducts struct {
	Data       []Product `json:"data"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}
