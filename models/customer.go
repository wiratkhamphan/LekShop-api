package models

import "time"

type Customer struct {
	Id         int       `json:"id"`
	CustomerID string    `json:"customer_id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
}
