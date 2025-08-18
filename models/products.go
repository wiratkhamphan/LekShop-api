package models

import "time"

type Product struct {
	ID          int       `json:"id"`
	ProductID   string    `json:"product_id"`
	Name        string    `json:"name"`
	Quantity    int       `json:"quantity"`
	CostPrice   float64   `json:"cost_price"`
	SellPrice   float64   `json:"sell_price"`
	Image       string    `json:"image"`
	Recommended bool      `json:"recommended"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
