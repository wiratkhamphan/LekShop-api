package controllers

import (
	"context"
	"dog/condb"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ProductPublic struct {
	ID              int      `json:"id"`
	SKU             string   `json:"product_id"`
	Name            string   `json:"name"`
	Brand           string   `json:"brand,omitempty"`
	Category        string   `json:"category,omitempty"`
	Gender          string   `json:"gender,omitempty"`
	Price           float64  `json:"price"` // = sell_price
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

// GET /products  (list + filter + sort + paginate)
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
	priceMin := c.Query("price_min")
	priceMax := c.Query("price_max")
	inStock := c.Query("in_stock") == "true"

	where := []string{"1=1"}
	args := []interface{}{}
	ai := 1

	if q != "" {
		where = append(where, "(p.name ILIKE $"+itoa(ai)+" OR p.brand ILIKE $"+itoa(ai)+" OR p.category ILIKE $"+itoa(ai)+")")
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
	if priceMin != "" {
		where = append(where, "p.sell_price >= $"+itoa(ai))
		if v, err := strconv.ParseFloat(priceMin, 64); err == nil {
			args = append(args, v)
			ai++
		}
	}
	if priceMax != "" {
		where = append(where, "p.sell_price <= $"+itoa(ai))
		if v, err := strconv.ParseFloat(priceMax, 64); err == nil {
			args = append(args, v)
			ai++
		}
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
  p.popularity_score,
  to_char(p.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF') AS created_at,
  p.quantity AS stock
FROM products p
WHERE ` + strings.Join(where, " AND ") + `
` // GROUP BY ไม่จำเป็นเพราะไม่มี aggregate

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

	// นับ total ตาม WHERE เดียวกัน
	countSQL := `SELECT COUNT(*) FROM products p WHERE ` + strings.Join(where, " AND ")
	if inStock {
		countSQL += " AND p.quantity > 0"
	}
	var total int
	if err := conn.QueryRow(context.Background(), countSQL, args[:ai-1]...).Scan(&total); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(ProductsListResp{Items: items, Total: total, Page: page, Limit: limit})
}

// GET /products/:id
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
  p.popularity_score,
  to_char(p.created_at, 'YYYY-MM-DD"T"HH24:MI:SSOF') AS created_at,
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

// GET /products/categories (facets)
func GetProductFacets(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	categories := splitCSV(c.Query("categories"))
	genders := splitCSV(c.Query("genders"))
	brands := splitCSV(c.Query("brand"))
	priceMin := c.Query("price_min")
	priceMax := c.Query("price_max")
	inStock := c.Query("in_stock") == "true"

	where := []string{"1=1"}
	args := []interface{}{}
	ai := 1

	if q != "" {
		where = append(where, "(name ILIKE $"+itoa(ai)+" OR brand ILIKE $"+itoa(ai)+" OR category ILIKE $"+itoa(ai)+")")
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
	if priceMin != "" {
		where = append(where, "sell_price >= $"+itoa(ai))
		if v, err := strconv.ParseFloat(priceMin, 64); err == nil {
			args = append(args, v)
			ai++
		}
	}
	if priceMax != "" {
		where = append(where, "sell_price <= $"+itoa(ai))
		if v, err := strconv.ParseFloat(priceMax, 64); err == nil {
			args = append(args, v)
			ai++
		}
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
	rows, err := conn.Query(context.Background(), `SELECT category, COUNT(*) FROM products `+w+` GROUP BY category ORDER BY COUNT(*) DESC, category ASC`, args...)
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
	rows, err = conn.Query(context.Background(), `SELECT brand, COUNT(*) FROM products `+w+` GROUP BY brand ORDER BY COUNT(*) DESC, brand ASC`, args...)
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
	rows, err = conn.Query(context.Background(), `SELECT gender, COUNT(*) FROM products `+w+` GROUP BY gender ORDER BY COUNT(*) DESC, gender ASC`, args...)
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
	_ = conn.QueryRow(context.Background(), `SELECT MIN(sell_price), MAX(sell_price) FROM products `+w, args...).Scan(&min, &max)
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
