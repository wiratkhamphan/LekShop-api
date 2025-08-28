package models

import "time"

type Product struct {
	ID            int        `json:"id"`
	ProductID     string     `json:"product_id"`
	Name          string     `json:"name"`
	Brand         *string    `json:"brand,omitempty"`
	Category      *string    `json:"category,omitempty"`
	Gender        *string    `json:"gender,omitempty"`
	Quantity      int        `json:"quantity"`
	CostPrice     *float64   `json:"cost_price,omitempty"`
	SellPrice     float64    `json:"sell_price"`
	OriginalPrice *float64   `json:"original_price,omitempty"`
	Image         *string    `json:"image,omitempty"`
	Recommended   bool       `json:"recommended"`
	Popularity    int        `json:"popularity_score"`
	Popular       bool       `json:"popular"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}
