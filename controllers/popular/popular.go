package popular

import (
	"context"
	"dog/condb"
	"dog/models"

	"github.com/gofiber/fiber/v2"
)

func Popular(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(
		context.Background(),
		"SELECT id, image_path, alt FROM popular_items WHERE slider_id=$1",
		1,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var items []models.PopularItem
	for rows.Next() {
		var item models.PopularItem
		if err := rows.Scan(&item.ID, &item.Image, &item.Alt); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		item.Image = "/static" + item.Image
		items = append(items, item)
	}

	return c.JSON(items)
}

// เพิ่ม Popular Item
func AddPopular(c *fiber.Ctx) error {
	// ดึง DB connection
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	// รับ alt จาก form
	alt := c.FormValue("alt")
	if alt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Alt text is required"})
	}

	// รับไฟล์
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Image is required"})
	}

	// บันทึกไฟล์ไป static folder
	savePath := "./static/images/slider/" + file.Filename
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Insert ลง DB
	_, err = conn.Exec(
		context.Background(),
		"INSERT INTO popular_items (slider_id, image_path, alt) VALUES ($1, $2, $3)",
		1, // slider_id
		"/images/slider/"+file.Filename,
		alt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Popular item added successfully"})
}

func UpdatePopular(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	// ดึง slider_id จาก URL param
	sliderID := c.Params("slider_id")

	// รับ JSON จาก body
	var item models.PopularItem
	if err := c.BodyParser(&item); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Update ข้อมูล
	_, err = conn.Exec(
		context.Background(),
		"UPDATE popular_items SET image_path=$1, alt=$2 WHERE slider_id=$3",
		item.Image,
		item.Alt,
		sliderID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Popular item updated successfully"})
}
