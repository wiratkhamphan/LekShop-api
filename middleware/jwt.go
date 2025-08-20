package middleware

import (
	"dog/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWTMiddleware(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing token"})
	}

	token := strings.TrimPrefix(auth, "Bearer ")
	userID, err := utils.ParseJWTToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
	}

	// เก็บ user_id เอาไว้ใช้ใน controller
	c.Locals("user_id", userID)
	return c.Next()
}
