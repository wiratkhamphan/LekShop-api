package models

import "time"

type Stock struct {
	ID        int       `json:"id"`
	ProductID string    `json:"product_id"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	CostPrice float64   `json:"cost_price"`
	SellPrice float64   `json:"sell_price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
