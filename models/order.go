package models

type CreateOrderItemReq struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Variant   string `json:"variant"`
}

type CreateOrderReq struct {
	Items         []CreateOrderItemReq `json:"items"`
	PaymentMethod string               `json:"payment_method"` // COD | BANK_TRANSFER | PROMPTPAY | CARD
}

type NextAction struct {
	Type        string `json:"type,omitempty"` // NONE | UPLOAD_SLIP | SHOW_PROMPTPAY | REDIRECT_GATEWAY
	URL         string `json:"url,omitempty"`
	QRImageURL  string `json:"qr_image_url,omitempty"`
	PayloadText string `json:"payload_text,omitempty"`
}

type CreateOrderResp struct {
	OrderID    int64       `json:"order_id"`
	Total      float64     `json:"total"`
	Message    string      `json:"message"`
	NextAction *NextAction `json:"next_action,omitempty"`
	Items      interface{} `json:"items,omitempty"`
}

type Order struct {
	ID            int64   `json:"id"`
	UserID        string  `json:"user_id"`
	Total         float64 `json:"total"`
	Status        string  `json:"status"`
	PaymentMethod string  `json:"payment_method"`
	PaymentStatus string  `json:"payment_status"`
	PaymentRef    *string `json:"payment_ref"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Variant   *string `json:"variant"`
}
