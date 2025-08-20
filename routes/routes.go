package routes

import (
	"dog/controllers"
	"dog/controllers/popular"
	"dog/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {

	//pos
	app.Post("/sales", controllers.CreateSale)
	app.Get("/sales", controllers.GetSales)
	app.Get("/sales/:sale_id", controllers.GetSaleByID)
	app.Put("/sales/:sale_id", controllers.UpdateSale)
	app.Delete("/sales/:sale_id", controllers.DeleteSale)

	//Emp
	app.Get("/employees", controllers.GetEmployees)
	app.Get("/employees/:employee_id", controllers.GetEmployeeByID)
	app.Post("/Next_EmployeeID", controllers.CreateEmployee)
	app.Put("/Employee/:emp_id", controllers.UpdateEmployee)

	// Logib
	app.Post("/Login", controllers.Login)
	app.Post("/LoginCustomer", controllers.LoginCustomer)

	// Stock & Products
	app.Get("/product", controllers.GetStock)
	app.Post("/product", controllers.AddStock)
	app.Put("/product/:product_id", controllers.UpdateStock)
	app.Patch("/product/:product_id/quantity", controllers.UpdateStockQuantity)
	app.Delete("/product/:product_id", controllers.DeleteStock)

	// Recommended Products
	app.Get("/products/recommended", controllers.GetRecommendedProducts)
	app.Patch("/products/:product_id/recommended", controllers.UpdateRecommended)

	//customers
	app.Post("/customers", controllers.CreateCustomer)
	app.Get("/customers", controllers.GetCustomers)
	app.Get("/customers/:customer_id", controllers.GetCustomerByID)
	app.Put("/customers/:customer_id", controllers.UpdateCustomer)

	// orders
	app.Post("/orders", middleware.JWTMiddleware, controllers.CreateOrder)
	app.Get("/orders", controllers.GetOrders)
	app.Get("/orders/:order_id", controllers.GetOrderByID)
	app.Put("/orders/:order_id", controllers.UpdateOrder)
	app.Delete("/orders/:order_id", controllers.DeleteOrder)

	//hero-slider
	app.Get("/hero-slider", popular.Popular)
	app.Post("/hero-slider", popular.AddPopular)
	app.Put("/hero-slider/:slider_id", popular.UpdatePopular)

}
