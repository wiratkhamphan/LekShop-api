package main

import (
	"dog/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {

	app := fiber.New()

	routes.RegisterRoutes(app)

	app.Listen(":8080")
}
