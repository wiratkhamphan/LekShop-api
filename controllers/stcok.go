package controllers

import (
	"context"
	"dog/condb"
	"dog/models"

	"github.com/gofiber/fiber/v2"
)

func GetStock(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect database",
		})
	}
	defer db.Close(context.Background())

	rows, err := db.Query(context.Background(), "SELECT id, product_id, name, quantity, cost_price, sell_price, created_at, updated_at FROM stock")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Query failed"})
	}

	var stocks []models.Stock
	for rows.Next() {
		var s models.Stock
		err := rows.Scan(&s.ID, &s.ProductID, &s.Name, &s.Quantity, &s.CostPrice, &s.SellPrice, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Scan failed"})
		}
		stocks = append(stocks, s)
	}

	return c.JSON(fiber.Map{
		"stocks": stocks,
	})
}

func AddStock(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect database",
		})
	}
	defer db.Close(context.Background())

	var input models.Stock
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	_, err = db.Exec(context.Background(),
		"INSERT INTO stock (product_id, name, quantity, cost_price, sell_price) VALUES ($1, $2, $3, $4, $5)",
		input.ProductID, input.Name, input.Quantity, input.CostPrice, input.SellPrice)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Insert failed"})
	}

	return c.JSON(fiber.Map{
		"message":   "Stock added successfully",
		"productID": input.ProductID,
		"name":      input.Name,
		"Quantity":  input.Quantity,
		"CostPrice": input.CostPrice,
		"SellPrice": input.SellPrice,
	})
}

func UpdateStockQuantity(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect database",
		})
	}
	defer db.Close(context.Background())

	productID := c.Params("product_id")
	type UpdateQty struct {
		Quantity int `json:"quantity"`
	}
	var input UpdateQty
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	_, err = db.Exec(context.Background(),
		"UPDATE stock SET quantity=$1, updated_at=NOW() WHERE product_id=$2",
		input.Quantity, productID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}

	return c.JSON(fiber.Map{
		"message":   "Stock quantity updated",
		"productID": productID,
		"quantity":  input.Quantity,
	})
}
