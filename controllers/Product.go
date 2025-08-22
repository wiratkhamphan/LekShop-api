package controllers

import (
	"context"
	"dog/condb"
	"dog/models"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ====================
// ดึงสินค้าทั้งหมด
// ====================
func GetStock(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	rows, err := db.Query(context.Background(),
		`SELECT id, product_id, name, quantity, cost_price, sell_price, image, recommended, created_at, updated_at 
		 FROM products`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Query failed"})
	}

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.ProductID, &p.Name, &p.Quantity, &p.CostPrice, &p.SellPrice, &p.Image, &p.Recommended, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Scan failed"})
		}
		products = append(products, p)
	}

	return c.JSON(fiber.Map{"products": products})
}

// ====================
// เพิ่มสินค้า (หรือ UPSERT) - รองรับ image ไม่บังคับ
// ====================
func AddStock(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	productID := c.FormValue("product_id")
	name := c.FormValue("name")
	brand := c.FormValue("brand")       // เพิ่มได้ (ไม่บังคับ)
	category := c.FormValue("category") // เพิ่มได้ (ไม่บังคับ)
	gender := c.FormValue("gender")     // men|women|unisex (ไม่บังคับ)

	quantity, _ := strconv.Atoi(c.FormValue("quantity"))
	costPrice, _ := strconv.ParseFloat(c.FormValue("cost_price"), 64)
	sellPrice, _ := strconv.ParseFloat(c.FormValue("sell_price"), 64)

	// ราคาป้าย (ถ้ามี)
	var originalPricePtr *float64
	if op := c.FormValue("original_price"); op != "" {
		if v, err := strconv.ParseFloat(op, 64); err == nil {
			originalPricePtr = &v
		}
	}

	recommended := c.FormValue("recommended") == "true"

	// รูป: "ไม่บังคับ"
	var imagePath *string
	if file, err := c.FormFile("image"); err == nil && file != nil {
		fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		savePath := "./static/images/products/" + fileName
		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		ip := "/static/images/products/" + fileName
		imagePath = &ip
	}

	// UPSERT (ถ้า image ไม่ส่ง -> ไม่ทับรูปเดิม)
	_, err = db.Exec(context.Background(), `
		INSERT INTO products
			(product_id, name, brand, category, gender, quantity, cost_price, sell_price, original_price, image, recommended, created_at, updated_at)
		VALUES
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11, now(), now())
		ON CONFLICT (product_id) DO UPDATE
		SET
			name            = EXCLUDED.name,
			brand           = EXCLUDED.brand,
			category        = EXCLUDED.category,
			gender          = EXCLUDED.gender,
			quantity        = EXCLUDED.quantity,
			cost_price      = EXCLUDED.cost_price,
			sell_price      = EXCLUDED.sell_price,
			original_price  = EXCLUDED.original_price,
			image           = COALESCE(EXCLUDED.image, products.image),
			recommended     = EXCLUDED.recommended,
			updated_at      = now()
			`,
		productID, name, brand, category, gender,
		quantity, costPrice, sellPrice, originalPricePtr, imagePath, recommended,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Insert failed: " + err.Error()})
	}

	product := models.Product{
		ProductID:     productID,
		Name:          name,
		Brand:         brand,
		Category:      category,
		Gender:        gender,
		Quantity:      quantity,
		CostPrice:     costPrice,
		SellPrice:     sellPrice,
		OriginalPrice: originalPricePtr,
		Image:         deref(imagePath),
		Recommended:   recommended,
	}

	return c.JSON(fiber.Map{"message": "Product added/updated", "product": product})
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ====================
// แก้ไขข้อมูลสินค้าเต็ม
// ====================
func UpdateStock(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	productID := c.Params("product_id")
	var input struct {
		Name        string  `json:"name"`
		Quantity    int     `json:"quantity"`
		CostPrice   float64 `json:"cost_price"`
		SellPrice   float64 `json:"sell_price"`
		Recommended bool    `json:"recommended"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	_, err = db.Exec(context.Background(),
		`UPDATE products
		 SET name=$1, quantity=$2, cost_price=$3, sell_price=$4, recommended=$5, updated_at=NOW()
		 WHERE product_id=$6`,
		input.Name, input.Quantity, input.CostPrice, input.SellPrice, input.Recommended, productID,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}

	return c.JSON(fiber.Map{"message": "Product updated", "productID": productID})
}

// ====================
// แก้ไขเฉพาะจำนวนสินค้า
// ====================
func UpdateStockQuantity(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	productID := c.Params("product_id")
	var input struct {
		Quantity int `json:"quantity"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	_, err = db.Exec(context.Background(),
		"UPDATE products SET quantity=$1, updated_at=NOW() WHERE product_id=$2",
		input.Quantity, productID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}

	return c.JSON(fiber.Map{
		"message":   "Stock quantity updated",
		"productID": productID,
		"quantity":  input.Quantity,
	})
}

// ====================
// ลบสินค้า
// ====================
func DeleteStock(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	productID := c.Params("product_id")
	_, err = db.Exec(context.Background(), "DELETE FROM products WHERE product_id=$1", productID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Delete failed"})
	}

	return c.JSON(fiber.Map{"message": "Product deleted", "productID": productID})
}

// ====================
// อัปเดตสถานะสินค้าแนะนำ
// ====================
func UpdateRecommended(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	productID := c.Params("product_id")
	var input struct {
		Recommended bool `json:"recommended"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	_, err = db.Exec(context.Background(),
		`UPDATE products SET recommended=$1, updated_at=NOW() WHERE product_id=$2`,
		input.Recommended, productID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}

	return c.JSON(fiber.Map{
		"message":     "Product recommendation updated",
		"productID":   productID,
		"recommended": input.Recommended,
	})
}

// ====================
// ดึงสินค้าแนะนำ (มีอยู่แล้ว)
// ====================
func GetRecommendedProducts(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	rows, err := db.Query(context.Background(),
		`SELECT id, product_id, name, quantity, cost_price, sell_price, image, recommended, created_at, updated_at
		 FROM products WHERE recommended = TRUE`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Query failed"})
	}

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.ProductID, &p.Name, &p.Quantity, &p.CostPrice, &p.SellPrice, &p.Image, &p.Recommended, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Scan failed"})
		}
		products = append(products, p)
	}

	return c.JSON(fiber.Map{"recommended_products": products})
}
