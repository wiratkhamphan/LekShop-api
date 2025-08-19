package controllers

import (
	"context"
	"dog/condb"
	"dog/models"
	"dog/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
)

// ✅ ดึงลูกค้าทั้งหมด
func GetCustomers(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), `
        SELECT id, customer_id, firstname, lastname, address, phone, email, created_at
        FROM customer
        ORDER BY id ASC
    `)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var cus models.Customer
		if err := rows.Scan(
			&cus.Id,
			&cus.CustomerID,
			&cus.FirstName,
			&cus.LastName,
			&cus.Address,
			&cus.Phone,
			&cus.Email,
			&cus.CreatedAt,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		customers = append(customers, cus)
	}

	return c.JSON(customers)
}

// ✅ ดึงลูกค้าด้วย customer_id
func GetCustomerByID(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	customerID := strings.TrimSpace(c.Params("customer_id"))
	fmt.Println("Param customerID:", customerID)

	var cus models.Customer
	err = conn.QueryRow(context.Background(), `
    SELECT customer_id, firstname, lastname, address, phone, email, created_at
    FROM customer
    WHERE customer_id = $1
`, customerID).Scan(
		&cus.CustomerID,
		&cus.FirstName,
		&cus.LastName,
		&cus.Address,
		&cus.Phone,
		&cus.Email,
		&cus.CreatedAt,
	)
	if err != nil {
		fmt.Println("Query error:", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Customer not found"})
	}

	fmt.Println("Customer found:", cus)

	return c.JSON(fiber.Map{"customer": cus})
}

// ✅ ฟังก์ชันสร้าง customer_id 6 หลัก
func GenerateNextCustomer(conn *pgx.Conn) (string, error) {
	var lastID string

	err := conn.QueryRow(
		context.Background(),
		`SELECT customer_id 
         FROM customer 
         ORDER BY customer_id::int DESC 
         LIMIT 1`,
	).Scan(&lastID)

	// ถ้ายังไม่มี customer เลย → เริ่มที่ 000001
	if err != nil {
		if err == pgx.ErrNoRows {
			return "000001", nil
		}
		return "", fmt.Errorf("failed to get last customer_id: %v", err)
	}

	num, err := strconv.Atoi(lastID)
	if err != nil {
		return "", fmt.Errorf("invalid customer_id format: %v", err)
	}

	newCusID := fmt.Sprintf("%06d", num+1) // เช่น 000002
	return newCusID, nil
}

// ✅ เพิ่มลูกค้าใหม่
func CreateCustomer(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	var cus models.Customer
	if err := c.BodyParser(&cus); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// hash password
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(cus.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}
	cus.Password = string(hashedPwd)

	// สร้าง customer_id 6 หลัก
	cus.CustomerID, err = GenerateNextCustomer(conn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	_, err = conn.Exec(context.Background(),
		`INSERT INTO customer (customer_id, firstname, lastname, email, password, phone, created_at)
         VALUES ($1,$2,$3,$4,$5,$6,NOW())`,
		cus.CustomerID, cus.FirstName, cus.LastName, cus.Email, cus.Password, cus.Phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":     "Customer created",
		"customer_id": cus.CustomerID,
	})
}

// ✅ อัปเดตข้อมูลลูกค้า
func UpdateCustomer(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	customerID := c.Params("customer_id")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "customer_id is required"})
	}

	var updateData models.Customer
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	commandTag, err := conn.Exec(context.Background(),
		`UPDATE customer 
         SET firstname=$1, lastname=$2, address=$3, phone=$4, email=$5, updated_at=NOW() 
         WHERE customer_id=$6`,
		updateData.FirstName,
		updateData.LastName,
		updateData.Address,
		updateData.Phone,
		updateData.Email,
		customerID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if commandTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Customer not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Customer updated successfully",
	})
}

// LoginCustomer handles login requests
func LoginCustomer(c *fiber.Ctx) error {
	var loginReq models.Login_Customer
	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	var cus models.Customer
	query := `SELECT customer_id, firstname, lastname, email, password FROM customer WHERE LOWER(email)=LOWER($1)`

	err = conn.QueryRow(context.Background(), query, loginReq.Email).Scan(
		&cus.CustomerID,
		&cus.FirstName,
		&cus.LastName,
		&cus.Email,
		&cus.Password,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Customer not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// ตรวจสอบ password
	if err := bcrypt.CompareHashAndPassword([]byte(cus.Password), []byte(loginReq.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Incorrect password"})
	}

	// สร้าง JWT token (ฟังก์ชัน utils.GenerateJWTToken)
	token, err := utils.GenerateJWTToken(cus.CustomerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Token generation failed"})
	}

	utils.SetJWTCookie(c, token)

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"customer": fiber.Map{
			"customer_id": cus.CustomerID,
			"firstname":   cus.FirstName,
			"lastname":    cus.LastName,
			"email":       cus.Email,
		},
		"token": token,
	})
}
