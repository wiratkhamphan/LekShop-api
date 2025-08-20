package models

import "time"

// ===== Enums / constants =====
const (
	// Order status (ตัวอย่างทั่วไป)
	OrderStatusPending    = "pending"
	OrderStatusPaid       = "paid"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusCancelled  = "cancelled"

	// Payment method
	PayMethodCOD          = "COD"
	PayMethodBankTransfer = "BANK_TRANSFER"
	PayMethodPromptPay    = "PROMPTPAY"
	PayMethodCard         = "CARD"

	// Payment status
	PayStatusPending = "pending"
	PayStatusReview  = "review"
	PayStatusPaid    = "paid"
	PayStatusFailed  = "failed"

	// Next action types
	NextNone          = "NONE"
	NextUploadSlip    = "UPLOAD_SLIP"
	NextShowPromptPay = "SHOW_PROMPTPAY"
	NextRedirect      = "REDIRECT_GATEWAY"
)

// ===== Requests =====

type CreateOrderItemReq struct {
	ProductID string  `json:"product_id"`        // required
	Quantity  int     `json:"quantity"`          // required, > 0
	Variant   *string `json:"variant,omitempty"` // optional; nil = ไม่ส่งมา, empty string = ระบุค่าว่าง
}

type CreateOrderReq struct {
	Items         []CreateOrderItemReq `json:"items"`                    // required
	PaymentMethod string               `json:"payment_method,omitempty"` // COD | BANK_TRANSFER | PROMPTPAY | CARD; ถ้าเว้นไว้ backend ตั้งค่า default ให้
}

// ===== Next Action =====

type NextAction struct {
	Type        string `json:"type,omitempty"`         // NONE | UPLOAD_SLIP | SHOW_PROMPTPAY | REDIRECT_GATEWAY
	URL         string `json:"url,omitempty"`          // ใช้ตอน REDIRECT_GATEWAY
	QRImageURL  string `json:"qr_image_url,omitempty"` // ใช้ตอน SHOW_PROMPTPAY
	PayloadText string `json:"payload_text,omitempty"` // ข้อความกำกับ/อธิบาย
}

// ===== Order / Items (DB Models / API Models) =====

type Order struct {
	ID            int64     `json:"id"`
	UserID        string    `json:"user_id"`
	Total         float64   `json:"total"` // แนะนำให้เป็น NUMERIC(12,2) ใน Postgres
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	PaymentStatus string    `json:"payment_status"`
	PaymentRef    *string   `json:"payment_ref,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"` // NUMERIC(12,2) ใน DB
	Quantity  int     `json:"quantity"`
	Variant   *string `json:"variant,omitempty"`
}

// ใช้ตอบกลับตอนสร้างออเดอร์ เพื่อให้ฝั่ง UI แสดงรายละเอียดได้สะดวก
type OrderLineResp struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Variant   *string `json:"variant,omitempty"`
	LineTotal float64 `json:"line_total"`
}

// ===== Create Order Response =====

type CreateOrderResp struct {
	OrderID    int64           `json:"order_id"`
	Total      float64         `json:"total"`
	Message    string          `json:"message"`
	NextAction *NextAction     `json:"next_action,omitempty"`
	Items      []OrderLineResp `json:"items,omitempty"` // แทน interface{} ให้เป็นโครงที่แน่นอน
}
