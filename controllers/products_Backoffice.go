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
)

type ProductPublic struct {
	ID              int      `json:"id"`
	SKU             string   `json:"product_id"`
	Name            string   `json:"name"`
	Brand           string   `json:"brand,omitempty"`
	Category        string   `json:"category,omitempty"`
	Gender          string   `json:"gender,omitempty"`
	Price           float64  `json:"price"`
	OriginalPrice   *float64 `json:"original_price,omitempty"`
	DiscountPercent int      `json:"discount_percent"`
	Image           string   `json:"image,omitempty"`
	Popularity      int      `json:"popularity_score"`
	CreatedAt       string   `json:"created_at"`
	Stock           int      `json:"stock"`
}

type ProductsListResp struct {
	Items []ProductPublic `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

// ====================
// เพิ่มสินค้าใหม่ (หรือแก้ไขถ้ามี product_id เดิม)
// ====================
func AddStock(c *fiber.Ctx) error {
	db, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to connect database"})
	}
	defer db.Close(context.Background())

	productID := c.FormValue("product_id")
	name := c.FormValue("name")
	brand := c.FormValue("brand")
	category := c.FormValue("category")
	gender := c.FormValue("gender")

	quantity, _ := strconv.Atoi(c.FormValue("quantity"))
	costPrice, _ := strconv.ParseFloat(c.FormValue("cost_price"), 64)
	sellPrice, _ := strconv.ParseFloat(c.FormValue("sell_price"), 64)

	// ราคาปกติ: "ไม่บังคับ"
	var originalPricePtr *float64
	if op := c.FormValue("original_price"); op != "" {
		if v, err := strconv.ParseFloat(op, 64); err == nil {
			originalPricePtr = &v
		}
	}

	recommended := c.FormValue("recommended") == "true"

	// อัปโหลดรูป (ถ้ามี)
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

	// Insert or Update
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
	// แปลง string เป็น *string (nil ถ้าว่าง)
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
// แก้ไขข้อมูลสินค้า
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
// แก้ไขจำนวนสินค้า (stock quantity)
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
// ดึงรายการสินค้า (พร้อมกรองและจัดเรียง)
// GET /products  (list + filter + sort + paginate)
// ====================

func GetProducts(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "16"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 60 {
		limit = 16
	}
	offset := (page - 1) * limit

	q := strings.TrimSpace(c.Query("q"))
	sort := c.Query("sort", "popularity") // popularity|newest|price_asc|price_desc
	categories := splitCSV(c.Query("categories"))
	genders := splitCSV(c.Query("genders"))
	brands := splitCSV(c.Query("brand"))
	priceMin := strings.TrimSpace(c.Query("price_min"))
	priceMax := strings.TrimSpace(c.Query("price_max"))
	inStock := c.Query("in_stock") == "true"

	where := []string{"1=1"}
	args := []interface{}{}
	ai := 1

	if q != "" {
		where = append(where,
			"(COALESCE(p.name,'') ILIKE $"+itoa(ai)+
				" OR COALESCE(p.brand,'') ILIKE $"+itoa(ai)+
				" OR COALESCE(p.category,'') ILIKE $"+itoa(ai)+")")
		args = append(args, "%"+q+"%")
		ai++
	}
	if len(categories) > 0 {
		where = append(where, "p.category = ANY($"+itoa(ai)+")")
		args = append(args, categories)
		ai++
	}
	if len(genders) > 0 {
		where = append(where, "p.gender = ANY($"+itoa(ai)+")")
		args = append(args, genders)
		ai++
	}
	if len(brands) > 0 {
		where = append(where, "p.brand = ANY($"+itoa(ai)+")")
		args = append(args, brands)
		ai++
	}
	// ✅ เพิ่มเงื่อนไขราคาเฉพาะเมื่อ parse สำเร็จ
	if v, ok := parseFloatSafe(priceMin); ok {
		where = append(where, "p.sell_price >= $"+itoa(ai))
		args = append(args, v)
		ai++
	}
	if v, ok := parseFloatSafe(priceMax); ok {
		where = append(where, "p.sell_price <= $"+itoa(ai))
		args = append(args, v)
		ai++
	}

	order := "p.popularity_score DESC, p.id DESC"
	switch sort {
	case "newest":
		order = "p.created_at DESC"
	case "price_asc":
		order = "p.sell_price ASC"
	case "price_desc":
		order = "p.sell_price DESC"
	}

	sql := `
SELECT
  p.id,
  p.product_id,
  p.name,
  p.brand,
  p.category,
  p.gender,
  p.sell_price AS price,
  p.original_price,
  CASE
    WHEN p.original_price IS NOT NULL AND p.original_price > p.sell_price
      THEN FLOOR((p.original_price - p.sell_price)/p.original_price*100)
    ELSE 0
  END AS discount_percent,
  COALESCE(p.image,'') AS image,
  p.popularity_score AS popularity, -- ✅ alias ให้ตรงฟิลด์สแกน
  to_char(p.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at, -- ✅ เวลามาตรฐาน
  p.quantity AS stock
FROM products p
WHERE ` + strings.Join(where, " AND ") + `
`
	if inStock {
		sql += " AND p.quantity > 0\n"
	}
	sql += " ORDER BY " + order + " OFFSET $" + itoa(ai) + " LIMIT $" + itoa(ai+1)

	args = append(args, offset, limit)

	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB connect failed"})
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), sql, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	items := make([]ProductPublic, 0, limit)
	for rows.Next() {
		var p ProductPublic
		if err := rows.Scan(
			&p.ID, &p.SKU, &p.Name, &p.Brand, &p.Category, &p.Gender,
			&p.Price, &p.OriginalPrice, &p.DiscountPercent, &p.Image,
			&p.Popularity, &p.CreatedAt, &p.Stock,
		); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		items = append(items, p)
	}

	// total
	countSQL := `SELECT COUNT(*) FROM products p WHERE ` + strings.Join(where, " AND ")
	if inStock {
		countSQL += " AND p.quantity > 0"
	}
	var total int
	// ใช้เฉพาะ args เงื่อนไข (ก่อน offset/limit) → args[:ai-1]
	if err := conn.QueryRow(context.Background(), countSQL, args[:ai-1]...).Scan(&total); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(ProductsListResp{Items: items, Total: total, Page: page, Limit: limit})
}

// ====================
// ดึงข้อมูลสินค้ารายตัว (by ID or product_id)
// GET /products/:id
// ====================

func GetProductByID(c *fiber.Ctx) error {
	id := c.Params("id")

	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB connect failed"})
	}
	defer conn.Close(context.Background())

	var p ProductPublic
	err = conn.QueryRow(context.Background(), `
SELECT
  p.id, p.product_id, p.name, p.brand, p.category, p.gender,
  p.sell_price AS price, p.original_price,
  CASE
    WHEN p.original_price IS NOT NULL AND p.original_price > p.sell_price
      THEN FLOOR((p.original_price - p.sell_price)/p.original_price*100)
    ELSE 0
  END AS discount_percent,
  COALESCE(p.image,'') AS image,
  p.popularity_score AS popularity,
  to_char(p.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at,
  p.quantity AS stock
FROM products p
WHERE p.product_id = $1 OR CAST(p.id AS TEXT) = $1
`, id).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Brand, &p.Category, &p.Gender,
		&p.Price, &p.OriginalPrice, &p.DiscountPercent, &p.Image,
		&p.Popularity, &p.CreatedAt, &p.Stock,
	)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(p)
}

// ====================
// ดึงข้อมูล facets สำหรับ filter
// GET /products/facets
// ====================
func GetProductFacets(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	categories := splitCSV(c.Query("categories"))
	genders := splitCSV(c.Query("genders"))
	brands := splitCSV(c.Query("brand"))
	priceMin := strings.TrimSpace(c.Query("price_min"))
	priceMax := strings.TrimSpace(c.Query("price_max"))
	inStock := c.Query("in_stock") == "true"

	where := []string{"1=1"}
	args := []interface{}{}
	ai := 1

	if q != "" {
		where = append(where,
			"(COALESCE(name,'') ILIKE $"+itoa(ai)+
				" OR COALESCE(brand,'') ILIKE $"+itoa(ai)+
				" OR COALESCE(category,'') ILIKE $"+itoa(ai)+")")
		args = append(args, "%"+q+"%")
		ai++
	}
	if len(categories) > 0 {
		where = append(where, "category = ANY($"+itoa(ai)+")")
		args = append(args, categories)
		ai++
	}
	if len(genders) > 0 {
		where = append(where, "gender = ANY($"+itoa(ai)+")")
		args = append(args, genders)
		ai++
	}
	if len(brands) > 0 {
		where = append(where, "brand = ANY($"+itoa(ai)+")")
		args = append(args, brands)
		ai++
	}
	if v, ok := parseFloatSafe(priceMin); ok {
		where = append(where, "sell_price >= $"+itoa(ai))
		args = append(args, v)
		ai++
	}
	if v, ok := parseFloatSafe(priceMax); ok {
		where = append(where, "sell_price <= $"+itoa(ai))
		args = append(args, v)
		ai++
	}
	if inStock {
		where = append(where, "quantity > 0")
	}

	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB connect failed"})
	}
	defer conn.Close(context.Background())

	w := "WHERE " + strings.Join(where, " AND ")

	type facet struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	out := fiber.Map{}

	// categories
	rows, err := conn.Query(context.Background(),
		`SELECT COALESCE(category,'') AS category, COUNT(*)
		 FROM products `+w+`
		 GROUP BY category
		 ORDER BY COUNT(*) DESC, category ASC`, args...)
	if err == nil {
		var list []facet
		for rows.Next() {
			var f facet
			if err := rows.Scan(&f.Name, &f.Count); err == nil && f.Name != "" {
				list = append(list, f)
			}
		}
		out["categories"] = list
		rows.Close()
	}

	// brands
	rows, err = conn.Query(context.Background(),
		`SELECT COALESCE(brand,'') AS brand, COUNT(*)
		 FROM products `+w+`
		 GROUP BY brand
		 ORDER BY COUNT(*) DESC, brand ASC`, args...)
	if err == nil {
		var list []facet
		for rows.Next() {
			var f facet
			if err := rows.Scan(&f.Name, &f.Count); err == nil && f.Name != "" {
				list = append(list, f)
			}
		}
		out["brands"] = list
		rows.Close()
	}

	// genders
	rows, err = conn.Query(context.Background(),
		`SELECT COALESCE(gender,'') AS gender, COUNT(*)
		 FROM products `+w+`
		 GROUP BY gender
		 ORDER BY COUNT(*) DESC, gender ASC`, args...)
	if err == nil {
		var list []facet
		for rows.Next() {
			var f facet
			if err := rows.Scan(&f.Name, &f.Count); err == nil && f.Name != "" {
				list = append(list, f)
			}
		}
		out["genders"] = list
		rows.Close()
	}

	// price range
	var min, max *float64
	_ = conn.QueryRow(context.Background(),
		`SELECT MIN(sell_price), MAX(sell_price) FROM products `+w, args...).
		Scan(&min, &max)
	out["price_range"] = fiber.Map{"min": min, "max": max}

	return c.JSON(out)
}

// helpers
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
func itoa(i int) string { return strconv.Itoa(i) }

// ====================
// อัปเดตสถานะสินค้าขายดี (popular)
// ====================
// ตัวอย่างการเรียก:
//
//	อัปเดตสินค้ารหัส "SKU1234" ให้เป็นสินค้าขายดี
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
