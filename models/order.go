package models

import "time"

// Order represents a customer order
// OrderID uses format ORD001 with incremental numbering
// Status can hold order state like pending, shipped, etc.
type Order struct {
	ID         int       `json:"id"`
	OrderID    string    `json:"order_id"`
	CustomerID string    `json:"customer_id"`
	TotalPrice float64   `json:"total_price"`
	OrderDate  time.Time `json:"order_date"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
