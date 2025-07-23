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

func GenerateNextSaleID(conn *pgx.Conn) (string, error) {
	var lastID string

	err := conn.QueryRow(context.Background(),
		`SELECT sale_id FROM sales ORDER BY id DESC LIMIT 1`,
	).Scan(&lastID)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return "SALE001", nil
		}
		return "", err
	}

	numPart := strings.TrimPrefix(lastID, "SALE")
	num, err := strconv.Atoi(numPart)
	if err != nil {
		return "", fmt.Errorf("invalid sale_id format in DB: %v", lastID)
	}

	newNum := num + 1
	newSaleID := fmt.Sprintf("SALE%03d", newNum)
	return newSaleID, nil
}

func CreateSale(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	newSaleID, err := GenerateNextSaleID(conn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var sale models.Sale
	if err := c.BodyParser(&sale); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	sale.SaleID = newSaleID

	// 1. เริ่ม Transaction เพื่อความปลอดภัยข้อมูล
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to start transaction"})
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	// 2. Insert ข้อมูลการขาย
	_, err = tx.Exec(context.Background(),
		`INSERT INTO sales (sale_id, employee_id, customer_id, product_id, quantity, total_price)
         VALUES ($1, $2, $3, $4, $5, $6)`,
		sale.SaleID, sale.EmployeeID, sale.CustomerID, sale.ProductID, sale.Quantity, sale.TotalPrice,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// 3. อัพเดต stock ลดจำนวนสินค้าคงเหลือ
	commandTag, err := tx.Exec(context.Background(),
		`UPDATE stock SET quantity = quantity - $1, updated_at = NOW() WHERE product_id = $2 AND quantity >= $1`,
		sale.Quantity, sale.ProductID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if commandTag.RowsAffected() == 0 {
		// rollback transaction เพราะสต็อกไม่พอหรือ product_id ไม่มี
		tx.Rollback(context.Background())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Insufficient stock or product not found"})
	}

	// 4. Commit transaction
	err = tx.Commit(context.Background())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Sale created and stock updated",
		"sale_id": newSaleID,
	})
}

func GetSales(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(),
		`SELECT id, sale_id, employee_id, customer_id, product_id, quantity, total_price, sale_date, created_at FROM sales ORDER BY id ASC`,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var sales []models.Sale
	for rows.Next() {
		var s models.Sale
		if err := rows.Scan(
			&s.ID, &s.SaleID, &s.EmployeeID, &s.CustomerID, &s.ProductID,
			&s.Quantity, &s.TotalPrice, &s.SaleDate, &s.CreatedAt,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		sales = append(sales, s)
	}

	return c.JSON(sales)
}

func GetSaleByID(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	saleID := c.Params("sale_id")

	var s models.Sale
	err = conn.QueryRow(context.Background(),
		`SELECT id, sale_id, employee_id, customer_id, product_id, quantity, total_price, sale_date, created_at
         FROM sales WHERE sale_id = $1`, saleID,
	).Scan(
		&s.ID, &s.SaleID, &s.EmployeeID, &s.CustomerID, &s.ProductID,
		&s.Quantity, &s.TotalPrice, &s.SaleDate, &s.CreatedAt,
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Sale not found"})
	}

	return c.JSON(s)
}

func UpdateSale(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	saleID := c.Params("sale_id")

	var updateData models.Sale
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	commandTag, err := conn.Exec(context.Background(),
		`UPDATE sales SET employee_id=$1, customer_id=$2, product_id=$3, quantity=$4, total_price=$5 WHERE sale_id=$6`,
		updateData.EmployeeID, updateData.CustomerID, updateData.ProductID, updateData.Quantity, updateData.TotalPrice, saleID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if commandTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Sale not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Sale updated successfully",
	})
}
func DeleteSale(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	saleID := c.Params("sale_id")

	commandTag, err := conn.Exec(context.Background(),
		`DELETE FROM sales WHERE sale_id=$1`, saleID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if commandTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Sale not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Sale deleted successfully",
	})
}
