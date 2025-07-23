package models

import "time"

type Employee struct {
	ID         int       `json:"id"`
	EmployeeID string    `json:"employee_id"`
	Password   string    `json:"password"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Position   string    `json:"position"`
	Salary     float64   `json:"salary"`
	HireDate   time.Time `json:"hire_date"`
	CreatedAt  time.Time `json:"created_at"`
}

type User_input struct {
	EmployeeID string `json:"employee_id"`
	Password   string `json:"password"`
}
