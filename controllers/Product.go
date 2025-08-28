package controllers

import (
	"context"
	"dog/condb"
	"dog/models"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

type ProductSearchResp struct {
	Items      []models.Product `json:"items"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	Total      int64            `json:"total"`
	TotalPages int              `json:"total_pages"`
}

// ====================
// ค้นหาสินค้า
// ====================

func SearchProducts(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	// ===== Params =====
	q := strings.TrimSpace(c.Query("q", ""))
	brand := strings.TrimSpace(c.Query("brand", ""))
	category := strings.TrimSpace(c.Query("category", ""))
	gender := strings.TrimSpace(c.Query("gender", "")) // men|women|unisex

	minPrice := c.Query("min_price", "")
	maxPrice := c.Query("max_price", "")

	recommended := c.Query("recommended", "") // "true"|"false"|""(ignore)
	popular := c.Query("popular", "")         // "true"|"false"|""(ignore)

	sort := strings.ToLower(c.Query("sort", "new")) // new|price_asc|price_desc|name|sold_desc
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 12)
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 60 {
		limit = 12
	}
	offset := (page - 1) * limit

	// ===== Build WHERE dynamically =====
	clauses := []string{"1=1"}
	args := []any{}
	arg := 1

	if q != "" {
		// ชื่อ/แบรนด์/หมวด ค้นหาแบบ ILIKE
		clauses = append(clauses, "(name ILIKE $"+strconv.Itoa(arg)+" OR brand ILIKE $"+strconv.Itoa(arg)+" OR category ILIKE $"+strconv.Itoa(arg)+")")
		args = append(args, "%"+q+"%")
		arg++
	}
	if brand != "" {
		clauses = append(clauses, "brand = $"+strconv.Itoa(arg))
		args = append(args, brand)
		arg++
	}
	if category != "" {
		clauses = append(clauses, "category = $"+strconv.Itoa(arg))
		args = append(args, category)
		arg++
	}
	if gender != "" {
		clauses = append(clauses, "gender = $"+strconv.Itoa(arg))
		args = append(args, gender)
		arg++
	}
	if minPrice != "" {
		clauses = append(clauses, "sell_price >= $"+strconv.Itoa(arg))
		args = append(args, toFloat(minPrice))
		arg++
	}
	if maxPrice != "" {
		clauses = append(clauses, "sell_price <= $"+strconv.Itoa(arg))
		args = append(args, toFloat(maxPrice))
		arg++
	}
	if recommended == "true" {
		clauses = append(clauses, "recommended = TRUE")
	} else if recommended == "false" {
		clauses = append(clauses, "recommended = FALSE")
	}
	if popular == "true" {
		clauses = append(clauses, "popular = TRUE")
	} else if popular == "false" {
		clauses = append(clauses, "popular = FALSE")
	}

	// ===== Sort =====
	orderBy := "updated_at DESC"
	switch sort {
	case "price_asc":
		orderBy = "sell_price ASC, updated_at DESC"
	case "price_desc":
		orderBy = "sell_price DESC, updated_at DESC"
	case "name":
		orderBy = "name ASC, updated_at DESC"
	case "sold_desc":
		// ถ้าต่อกับตารางยอดขายในอนาคต ให้เติม LEFT JOIN + ORDER BY sold DESC
		orderBy = "updated_at DESC"
	default:
		orderBy = "updated_at DESC"
	}

	where := strings.Join(clauses, " AND ")

	// ===== Query with window count =====
	sql := `
		SELECT
			id, product_id, name, brand, category, gender,
			quantity, cost_price, sell_price, original_price,
			COALESCE(image,'') as image,
			recommended, popular, created_at, updated_at,
			COUNT(*) OVER() AS total_count
		FROM products
		WHERE ` + where + `
		ORDER BY ` + orderBy + `
		LIMIT $` + strconv.Itoa(arg) + ` OFFSET $` + strconv.Itoa(arg+1)

	args = append(args, limit, offset)

	rows, qerr := db.Query(context.Background(), sql, args...)
	if qerr != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Query failed: " + qerr.Error()})
	}
	defer rows.Close()

	var items []models.Product
	var total int64 = 0
	for rows.Next() {
		var p models.Product
		var t int64
		if err := rows.Scan(
			&p.ID, &p.ProductID, &p.Name, &p.Brand, &p.Category, &p.Gender,
			&p.Quantity, &p.CostPrice, &p.SellPrice, &p.OriginalPrice,
			&p.Image, &p.Recommended, &p.Popular, &p.CreatedAt, &p.UpdatedAt, &t,
		); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Scan failed"})
		}
		total = t
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Rows error"})
	}

	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return c.JSON(ProductSearchResp{
		Items:      items,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

func toFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// ====================
// ดึงสินค้าทั้งหมด (Backoffice)
// ====================
func GetStock(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	rows, err := db.Query(context.Background(),
		`SELECT id, product_id, name, brand, category, gender, quantity, cost_price, sell_price, original_price, image, recommended, popular, created_at, updated_at FROM products`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Query failed"})
	}

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.ProductID, &p.Name, &p.Brand, &p.Category, &p.Gender, &p.Quantity, &p.CostPrice, &p.SellPrice, &p.OriginalPrice, &p.Image, &p.Recommended, &p.Popular, &p.CreatedAt, &p.UpdatedAt); err != nil {
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

	// Helper: convert string to *string
	toPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}

	product := models.Product{
		ProductID:     productID,
		Name:          name,
		Brand:         toPtr(brand),
		Category:      toPtr(category),
		Gender:        toPtr(gender),
		Quantity:      quantity,
		CostPrice:     &costPrice,
		SellPrice:     sellPrice,
		OriginalPrice: originalPricePtr,
		Image:         imagePath,
		Recommended:   recommended,
	}

	return c.JSON(fiber.Map{"message": "Product added/updated", "product": product})
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

	limit := c.QueryInt("limit", 0)

	baseQuery := `
        SELECT id, product_id, name, quantity, cost_price, sell_price, image, recommended, created_at, updated_at
        FROM products
        WHERE recommended = TRUE
        ORDER BY updated_at DESC`
	var rows pgx.Rows
	if limit > 0 {
		baseQuery += " LIMIT $1"
		rows, err = db.Query(context.Background(), baseQuery, limit)
	} else {
		rows, err = db.Query(context.Background(), baseQuery)
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Query failed"})
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.ProductID, &p.Name, &p.Quantity,
			&p.CostPrice, &p.SellPrice, &p.Image, &p.Recommended, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Scan failed"})
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Rows error"})
	}

	return c.JSON(fiber.Map{"recommended_products": products})
}

// ====================
// ดึงสินค้ารุ่นยอดนิยม (manual/auto)
//
//	GET /popular?mode=auto|manual&days=30&limit=12
//
// ====================
func GetPopularProducts(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	mode := c.Query("mode", "auto") // auto | manual
	days := c.QueryInt("days", 30)  // นับจากวันนี้ย้อนหลัง X วัน (เฉพาะโหมด auto)
	limit := c.QueryInt("limit", 12)

	// ปรับสถานะออเดอร์ตามระบบคุณ เช่น paid/shipped/completed
	validStatuses := []string{"paid", "shipped", "completed"}

	type PopularItem struct {
		ID          int       `json:"id"`
		ProductID   string    `json:"product_id"`
		Name        string    `json:"name"`
		SellPrice   float64   `json:"sell_price"`
		Image       *string   `json:"image"`
		Recommended bool      `json:"recommended"`
		Quantity    int       `json:"quantity"`
		UpdatedAt   time.Time `json:"updated_at"`
		SoldInRange int64     `json:"sold_in_range"` // ใช้เฉพาะ mode=auto
	}

	var rows pgx.Rows

	switch mode {
	case "manual":
		// ต้องมีคอลัมน์ products.popular = true
		query := `
			SELECT id, product_id, name, sell_price, image, recommended, quantity, updated_at, 0 as sold_in_range
			FROM products
			WHERE popular = TRUE
			ORDER BY updated_at DESC`
		if limit > 0 {
			query += " LIMIT $1"
			rows, err = db.Query(context.Background(), query, limit)
		} else {
			rows, err = db.Query(context.Background(), query)
		}
	default:
		// auto: นับยอดขายภายใน N วันย้อนหลัง
		// ปรับชื่อคอลัมน์/ตารางตามจริง
		base := `
			WITH oi AS (
				SELECT
					oi.product_id,
					SUM(oi.quantity) AS sold
				FROM order_items oi
				JOIN orders o ON o.id = oi.order_id
				WHERE o.created_at >= (NOW() - ($1::int || ' days')::interval)
				  AND o.status = ANY($2)
				GROUP BY oi.product_id
			)
			SELECT
				p.id, p.product_id, p.name, p.sell_price, p.image, p.recommended, p.quantity, p.updated_at,
				COALESCE(oi.sold, 0) AS sold_in_range
			FROM products p
			LEFT JOIN oi ON oi.product_id = p.product_id
			ORDER BY oi.sold DESC NULLS LAST, p.updated_at DESC`
		if limit > 0 {
			base += " LIMIT $3"
			rows, err = db.Query(context.Background(), base, days, validStatuses, limit)
		} else {
			rows, err = db.Query(context.Background(), base, days, validStatuses)
		}
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Query failed: " + err.Error()})
	}
	defer rows.Close()

	var items []PopularItem
	for rows.Next() {
		var it PopularItem
		if err := rows.Scan(
			&it.ID, &it.ProductID, &it.Name, &it.SellPrice, &it.Image,
			&it.Recommended, &it.Quantity, &it.UpdatedAt, &it.SoldInRange,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Scan failed"})
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Rows error"})
	}

	// ส่งออกแบบ array โดยตรง หรือจะห่อเป็น {items: []} ก็ได้
	return c.JSON(items)
}

// ====================
// อัปเดตสถานะ popular (manual)
//
//	PATCH /products/:product_id/popular
//	body: { "popular": true }
//
// ====================
func UpdatePopularFlag(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	productID := c.Params("product_id")
	var input struct {
		Popular bool `json:"popular"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	_, err = db.Exec(context.Background(),
		`UPDATE products SET popular=$1, updated_at=NOW() WHERE product_id=$2`,
		input.Popular, productID,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Update failed"})
	}

	return c.JSON(fiber.Map{
		"message":   "Product popular flag updated",
		"productID": productID,
		"popular":   input.Popular,
	})
}
