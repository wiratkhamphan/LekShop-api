package main

import (
	"dog/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	// ✅ CORS: อนุญาตทั้ง
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin, Content-Type, Authorization, Idempotency-Key, X-Requested-With",
		AllowCredentials: false, // ถ้าจะใช้คุกกี้ ค่อยเปลี่ยนเป็น true และต้องคง AllowOrigins แบบระบุโดเมน (ห้าม *)
		// ExposeHeaders:  "X-Request-Id", // ถ้าต้องการอ่าน header ตอบกลับพิเศษให้เพิ่มที่นี่
	}))

	// Static files (ถ้ามี)
	app.Static("/static", "./static")

	// API routes
	routes.RegisterRoutes(app)

	app.Listen(":8080")
}
