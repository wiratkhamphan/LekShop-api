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

	origins := os.Getenv("ALLOW_ORIGINS")
	var allowList []string

	if strings.TrimSpace(origins) == "" {
		// default dev
		allowList = []string{"http://localhost:3000", "http://127.0.0.1:5500", "https://lek-shop.vercel.app"}
	} else {
		allowList = strings.Split(origins, ",")
	}

	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			for _, o := range allowList {
				if strings.TrimSpace(o) == origin {
					return true
				}
			}
			return false
		},
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Set-Cookie",
		AllowCredentials: true,
	}))

	app.Static("/static", "./static")

	routes.RegisterRoutes(app)

	log.Fatal(app.Listen(":8080"))
}
