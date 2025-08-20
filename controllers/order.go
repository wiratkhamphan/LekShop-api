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

// GenerateNextOrderID creates the next incremental order ID using ORD prefix
func GenerateNextOrderID(conn *pgx.Conn) (string, error) {
	var lastID string

	err := conn.QueryRow(context.Background(),
		`SELECT order_id FROM orders ORDER BY id DESC LIMIT 1`,
	).Scan(&lastID)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return "ORD001", nil
		}
		return "", err
	}

	numPart := strings.TrimPrefix(lastID, "ORD")
	num, err := strconv.Atoi(numPart)
	if err != nil {
		return "", fmt.Errorf("invalid order_id format in DB: %v", lastID)
	}

	newNum := num + 1
	newOrderID := fmt.Sprintf("ORD%03d", newNum)
	return newOrderID, nil
}

// CreateOrder inserts a new order record
func CreateOrder(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	newOrderID, err := GenerateNextOrderID(conn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var order models.Order
	if err := c.BodyParser(&order); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	order.OrderID = newOrderID

	_, err = conn.Exec(context.Background(),
		`INSERT INTO orders (order_id, customer_id, total_price, status) VALUES ($1,$2,$3,$4)`,
		order.OrderID, order.CustomerID, order.TotalPrice, order.Status,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Order created",
		"order_id": newOrderID,
	})
}

// GetOrders returns all orders
func GetOrders(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(),
		`SELECT id, order_id, customer_id, total_price, order_date, status, created_at FROM orders ORDER BY id ASC`,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.OrderID, &o.CustomerID, &o.TotalPrice, &o.OrderDate, &o.Status, &o.CreatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		orders = append(orders, o)
	}

	return c.JSON(orders)
}

// GetOrderByID returns a single order by its ID
func GetOrderByID(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	orderID := c.Params("order_id")

	var o models.Order
	err = conn.QueryRow(context.Background(),
		`SELECT id, order_id, customer_id, total_price, order_date, status, created_at FROM orders WHERE order_id = $1`, orderID,
	).Scan(&o.ID, &o.OrderID, &o.CustomerID, &o.TotalPrice, &o.OrderDate, &o.Status, &o.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	return c.JSON(o)
}

// UpdateOrder updates order information
func UpdateOrder(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	orderID := c.Params("order_id")

	var updateData models.Order
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	commandTag, err := conn.Exec(context.Background(),
		`UPDATE orders SET customer_id=$1, total_price=$2, status=$3 WHERE order_id=$4`,
		updateData.CustomerID, updateData.TotalPrice, updateData.Status, orderID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if commandTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	return c.JSON(fiber.Map{"message": "Order updated successfully"})
}

// DeleteOrder removes an order by ID
func DeleteOrder(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	orderID := c.Params("order_id")

	commandTag, err := conn.Exec(context.Background(),
		`DELETE FROM orders WHERE order_id=$1`, orderID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if commandTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Order not found"})
	}

	return c.JSON(fiber.Map{"message": "Order deleted successfully"})
}
