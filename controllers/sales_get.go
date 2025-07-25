package controllers

import (
	"context"
	"dog/condb"
	"dog/models"

	"github.com/gofiber/fiber/v2"
)

func SalesGet(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect database",
		})
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(),
		"SELECT id, employee_id, customer_id, product_id, quantity, total_price, Sale_date FROM sales")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer rows.Close()

	var sales []fiber.Map

	for rows.Next() {
		var sale models.Sale
		err := rows.Scan(
			&sale.ID,
			&sale.EmployeeID,
			&sale.CustomerID,
			&sale.ProductID,
			&sale.Quantity,
			&sale.TotalPrice,
			&sale.SaleDate,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		sales = append(sales, fiber.Map{
			"id":          sale.ID,
			"employee_id": sale.EmployeeID,
			"customer_id": sale.CustomerID,
			"Product_id":  sale.ProductID,
			"Quantity":    sale.Quantity,
			"TotalPrice":  sale.TotalPrice,
		})
	}

	return c.Status(fiber.StatusOK).JSON(sales)
}
func GET_sale_by_id(c *fiber.Ctx) error {
	in_emp_id := c.Params("in_id")

	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect database",
		})
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(
		context.Background(),
		"SELECT id, employee_id, customer_id, product_id, quantity, total_price, sale_date FROM sales WHERE employee_id = $1",
		in_emp_id,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	defer rows.Close()

	var sales []fiber.Map

	for rows.Next() {
		var sale models.Sale
		err := rows.Scan(
			&sale.ID,
			&sale.EmployeeID,
			&sale.CustomerID,
			&sale.ProductID,
			&sale.Quantity,
			&sale.TotalPrice,
			&sale.SaleDate,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		sales = append(sales, fiber.Map{
			"id":          sale.ID,
			"employee_id": sale.EmployeeID,
			"customer_id": sale.CustomerID,
			"product_id":  sale.ProductID,
			"quantity":    sale.Quantity,
			"total_price": sale.TotalPrice,
			"sale_date":   sale.SaleDate,
		})
	}

	if len(sales) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No sales found for employee_id: " + in_emp_id,
		})
	}

	return c.Status(fiber.StatusOK).JSON(sales)
}
