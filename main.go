package main

import (
	"dog/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {

	app := fiber.New()

	// Setup CORS middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Set("Access-Control-Allow-Credentials", "true")

		if c.Method() == "OPTIONS" {
			c.SendStatus(fiber.StatusNoContent)
			return nil
		}

		return c.Next()
	})
	app.Static("/static", "./static")
	routes.RegisterRoutes(app)

	app.Listen(":8080")
}
