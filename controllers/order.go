package controllers

import (
	"context"
	"dog/condb"
	"dog/models"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

func paymentStatusInitial(method string) string {
	switch strings.ToUpper(method) {
	case "COD":
		return "pending"
	case "BANK_TRANSFER":
		return "review"
	case "PROMPTPAY":
		return "pending"
	case "CARD":
		return "pending"
	default:
		return "pending"
	}
}

func CreateOrder(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req models.CreateOrderReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "items is empty"})
	}
	if req.PaymentMethod == "" {
		req.PaymentMethod = "COD"
	}

	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	ctx := context.Background()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "begin tx failed"})
	}
	defer tx.Rollback(ctx)

	type calc struct {
		ProductID string
		Name      string
		Price     float64
		Qty       int
		Variant   string
		Line      float64
	}

	var lines []calc
	var grand float64

	for _, it := range req.Items {
		if it.ProductID == "" || it.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid item"})
		}
		var (
			name      string
			stockQty  int
			sellPrice float64
		)
		row := tx.QueryRow(ctx, `
			SELECT name, quantity, sell_price
			FROM products
			WHERE product_id = $1
			FOR UPDATE
		`, it.ProductID)
		if err := row.Scan(&name, &stockQty, &sellPrice); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": fmt.Sprintf("product not found: %s", it.ProductID)})
		}
		if stockQty < it.Quantity {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":       "insufficient stock",
				"product_id":  it.ProductID,
				"stock_left":  stockQty,
				"request_qty": it.Quantity,
			})
		}
		line := float64(it.Quantity) * sellPrice
		grand += line
		lines = append(lines, calc{
			ProductID: it.ProductID,
			Name:      name,
			Price:     sellPrice,
			Qty:       it.Quantity,
			Variant:   *it.Variant,
			Line:      line,
		})
	}

	var orderID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO orders (user_id, total, status, payment_method, payment_status)
		VALUES ($1, $2, 'pending', $3, $4)
		RETURNING id
	`, userID, grand, req.PaymentMethod, paymentStatusInitial(req.PaymentMethod)).Scan(&orderID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "create order failed"})
	}

	for _, l := range lines {
		if _, err := tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, name, price, quantity, variant)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6,''))
		`, orderID, l.ProductID, l.Name, l.Price, l.Qty, l.Variant); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "insert order_items failed"})
		}
		if _, err := tx.Exec(ctx, `
			UPDATE stock SET quantity = quantity - $1, updated_at = NOW()
			WHERE product_id = $2
		`, l.Qty, l.ProductID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "update stock failed"})
		}
	}

	var next *models.NextAction
	switch strings.ToUpper(req.PaymentMethod) {
	case "BANK_TRANSFER":
		next = &models.NextAction{Type: "UPLOAD_SLIP"}
	case "PROMPTPAY":
		next = &models.NextAction{
			Type:        "SHOW_PROMPTPAY",
			QRImageURL:  fmt.Sprintf("/api/orders/%d/promptpay-qr.png", orderID),
			PayloadText: fmt.Sprintf("PromptPay สำหรับออเดอร์ #%d ยอดชำระ %.2f บาท", orderID, grand),
		}
	case "CARD":
		next = &models.NextAction{
			Type: "REDIRECT_GATEWAY",
			URL:  fmt.Sprintf("https://example-gateway.test/pay?order_id=%d&amount=%.2f", orderID, grand),
		}
	default:
		next = &models.NextAction{Type: "NONE"}
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "commit failed"})
	}

	return c.JSON(models.CreateOrderResp{
		OrderID:    orderID,
		Total:      grand,
		Message:    "สร้างคำสั่งซื้อสำเร็จ",
		NextAction: next,
	})
}

func GetOrders(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	userID := c.Query("user_id", "")

	ctx := context.Background()

	query := `
		SELECT id, user_id, total, status, payment_method, payment_status, payment_ref, created_at, updated_at
		FROM orders`
	var rows pgx.Rows // <- ถ้าคุณ import "github.com/jackc/pgx/v4"
	var qerr error

	if userID != "" {
		query += ` WHERE user_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3`
		rows, qerr = conn.Query(ctx, query, userID, limit, offset)
	} else {
		query += ` ORDER BY id DESC LIMIT $1 OFFSET $2`
		rows, qerr = conn.Query(ctx, query, limit, offset)
	}
	if qerr != nil {
		return c.Status(500).JSON(fiber.Map{"error": qerr.Error()})
	}
	defer rows.Close()

	var list []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.PaymentMethod, &o.PaymentStatus, &o.PaymentRef, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		list = append(list, o)
	}
	return c.JSON(fiber.Map{
		"items":  list,
		"limit":  limit,
		"offset": offset,
	})
}

func GetOrderByID(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	id := c.Params("order_id")
	ctx := context.Background()

	var o models.Order
	err = conn.QueryRow(ctx, `
		SELECT id, user_id, total, status, payment_method, payment_status, payment_ref, created_at, updated_at
		FROM orders WHERE id = $1
	`, id).Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.PaymentMethod, &o.PaymentStatus, &o.PaymentRef, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "order not found"})
	}

	rows, err := conn.Query(ctx, `
		SELECT id, order_id, product_id, name, price, quantity, variant
		FROM order_items WHERE order_id = $1
	`, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var it models.OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.Name, &it.Price, &it.Quantity, &it.Variant); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		items = append(items, it)
	}

	return c.JSON(fiber.Map{
		"order": o,
		"items": items,
	})
}

type updateOrderReq struct {
	Status        *string `json:"status"`
	PaymentStatus *string `json:"payment_status"`
	PaymentRef    *string `json:"payment_ref"`
}

func UpdateOrder(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	id := c.Params("order_id")
	var req updateOrderReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	ctx := context.Background()
	sets := []string{}
	args := []interface{}{}
	i := 1

	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", i))
		args = append(args, *req.Status)
		i++
	}
	if req.PaymentStatus != nil {
		sets = append(sets, fmt.Sprintf("payment_status = $%d", i))
		args = append(args, *req.PaymentStatus)
		i++
	}
	if req.PaymentRef != nil {
		sets = append(sets, fmt.Sprintf("payment_ref = $%d", i))
		args = append(args, *req.PaymentRef)
		i++
	}

	if len(sets) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "no fields to update"})
	}
	sets = append(sets, "updated_at = NOW()")

	args = append(args, id)
	query := fmt.Sprintf("UPDATE orders SET %s WHERE id = $%d", strings.Join(sets, ","), i)

	if _, err := conn.Exec(ctx, query, args...); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "updated"})
}

func DeleteOrder(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	id := c.Params("order_id")
	ctx := context.Background()

	res, err := conn.Exec(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if res.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "order not found"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}
