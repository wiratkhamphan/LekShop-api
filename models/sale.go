package models

import "time"

type Sale struct {
	ID         int       `json:"id"`
	SaleID     string    `json:"sale_id"`
	EmployeeID string    `json:"employee_id"`
	CustomerID string    `json:"customer_id"`
	ProductID  string    `json:"product_id"`
	Quantity   int       `json:"quantity"`
	TotalPrice float64   `json:"total_price"`
	SaleDate   time.Time `json:"sale_date"`
	CreatedAt  time.Time `json:"created_at"`
}
