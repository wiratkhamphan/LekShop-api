package main

import (
	"log"
	"os"
	"strings"

	"dog/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	allow := os.Getenv("ALLOW_ORIGINS")
	if strings.TrimSpace(allow) == "" {
		// ค่า default สำหรับ dev
		allow = "http://127.0.0.1:5500,http://localhost:5500,http://localhost:3000"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allow, // คั่นด้วย comma
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Set-Cookie",
		AllowCredentials: true,
	}))

	app.Static("/static", "./static")

	routes.RegisterRoutes(app)

	log.Fatal(app.Listen(":8080"))
}
