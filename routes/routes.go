package routes

import (
	"dog/controllers"
	"dog/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	// POS
	app.Post("/sales", controllers.CreateSale)
	app.Get("/sales", controllers.GetSales)
	app.Get("/sales/:sale_id", controllers.GetSaleByID)
	app.Put("/sales/:sale_id", controllers.UpdateSale)
	app.Delete("/sales/:sale_id", controllers.DeleteSale)

	// Login
	app.Post("/Login", controllers.Login)
	app.Post("/LoginCustomer", controllers.LoginCustomer)

	// Public Product Catalog (ลูกค้า)
	app.Get("/products", controllers.GetProducts)
	app.Get("/products/categories", controllers.GetProductFacets)
	app.Get("/products/recommended", controllers.GetRecommendedProducts)
	app.Get("/products/:id", controllers.GetProductByID)
	app.Get("/api/products", controllers.SearchProducts)
	app.Get("/popular", controllers.GetPopularProducts)

	// Customers (ลูกค้า)
	app.Post("/customers", controllers.CreateCustomer)
	app.Get("/customers", controllers.GetCustomers)
	app.Get("/customers/:customer_id", controllers.GetCustomerByID)
	app.Put("/customers/:customer_id", controllers.UpdateCustomer)

	// Orders (ลูกค้า)
	app.Post("/orders", middleware.JWTMiddleware, controllers.CreateOrder)
	app.Get("/orders", controllers.GetOrders)
	app.Get("/orders/:order_id", controllers.GetOrderByID)
	app.Put("/orders/:order_id", controllers.UpdateOrder)
	app.Delete("/orders/:order_id", controllers.DeleteOrder)

	// ===== Admin/Backoffice API Group =====
	admin := app.Group("/admin", middleware.JWTMiddleware)

	// Stock & Products (หลังบ้าน)
	admin.Get("/product", controllers.GetStock)
	admin.Post("/product", controllers.AddStock)
	admin.Put("/product/:product_id", controllers.UpdateStock)
	admin.Patch("/product/:product_id/quantity", controllers.UpdateStockQuantity)
	admin.Delete("/product/:product_id", controllers.DeleteStock)
	admin.Patch("/products/:product_id/popular", controllers.UpdatePopularFlag)

	// Employees (หลังบ้าน)
	admin.Get("/employees", controllers.GetEmployees)
	admin.Get("/employees/:employee_id", controllers.GetEmployeeByID)
	admin.Post("/Next_EmployeeID", controllers.CreateEmployee)
	admin.Put("/Employee/:emp_id", controllers.UpdateEmployee)

	// Orders (หลังบ้าน)
	admin.Get("/orders", controllers.GetOrders)
	admin.Get("/orders/:order_id", controllers.GetOrderByID)
	admin.Put("/orders/:order_id", controllers.UpdateOrder)
	admin.Delete("/orders/:order_id", controllers.DeleteOrder)

	// Customers (หลังบ้าน)
	admin.Get("/customers", controllers.GetCustomers)
	admin.Get("/customers/:customer_id", controllers.GetCustomerByID)
	admin.Put("/customers/:customer_id", controllers.UpdateCustomer)

	// ===== Debug routes (เปิดใช้ชั่วคราวเวลาตามหา 404) =====
	// for _, r := range app.GetRoutes() {
	// 	println(r.Method, r.Path)
	// }
	// app.Use(func(c *fiber.Ctx) error {
	// 	fmt.Printf("NotFound -> %s %s\n", c.Method(), c.Path())
	// 	return c.Status(404).JSON(fiber.Map{"error": "not found"})
	// })
}
