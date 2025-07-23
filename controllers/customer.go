package controllers

import (
	"context"
	"dog/condb"
	"dog/models"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

func GetCustomers(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), `
        SELECT id, customer_id, name, address, phone, email, created_at
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
		if err := rows.Scan(&cus.Id, &cus.CustomerID, &cus.Name, &cus.Address, &cus.Phone, &cus.Email, &cus.CreatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		customers = append(customers, cus)
	}

	return c.JSON(customers)
}

func GetCustomerByID(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	customerID := c.Params("customer_id")

	var cus models.Customer
	err = conn.QueryRow(context.Background(), `
        SELECT id, customer_id, name, address, phone, email, created_at
        FROM customer
        WHERE customer_id = $1
    `, customerID).Scan(&cus.Id, &cus.CustomerID, &cus.Name, &cus.Address, &cus.Phone, &cus.Email, &cus.CreatedAt)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Customer not found"})
	}

	return c.JSON(cus)
}

func GenerateNextCustomer(conn *pgx.Conn) (string, error) {
	var UseId string

	err := conn.QueryRow(
		context.Background(),
		`SELECT customer_id 
		 FROM customer 
		 WHERE customer_id LIKE 'CUS%' 
		 ORDER BY customer_id DESC 
		 LIMIT 1`,
	).Scan(&UseId)

	if err != nil {
		if err == pgx.ErrNoRows {
			return "USE000", nil
		}
		return "", nil
	}

	numStr := strings.TrimPrefix(UseId, "USE")
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return "", fmt.Errorf("invalid customer_id format: %v", err)
	}

	newCusID := fmt.Sprintf("USE%03d", num+1)

	return newCusID, nil
}

func CreateCustomer(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	newUseID, err := GenerateNextEmployeeID(conn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var use models.Customer
	if err := c.BodyParser(&use); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	use.CustomerID = newUseID

	_, err = conn.Exec(context.Background(),
		`INSERT INTO customer (customer_id, name, address, phone, email, created_at)
     VALUES ($1, $2, $3, $4, $5, CURRENT_DATE)`,
		use.CustomerID,
		use.Name,
		use.Address,
		use.Phone,
		use.Email,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "customer created",
		"customer_id": newUseID,
	})
}

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
		`UPDATE customer SET name=$1, address=$2, phone=$3, email=$4, updated_at=NOW() WHERE customer_id=$5`,
		updateData.Name,
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
