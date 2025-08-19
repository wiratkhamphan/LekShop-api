package models

import "time"

type Customer struct {
	Id         int       `json:"id"`
	CustomerID string    `json:"customer_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Address    *string   `json:"address"` // รองรับ NULL
	Phone      *string   `json:"phone"`   // รองรับ NULL
	Email      string    `json:"email"`
	Password   string    `json:"password"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Login_Customer struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
