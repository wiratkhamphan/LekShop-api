package routes

import (
	"dog/controllers"
	"dog/controllers/popular"

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

	//stock
	app.Get("/stock", controllers.GetStock)
	app.Post("/stock", controllers.AddStock)
	app.Put("/stock/:product_id", controllers.UpdateStockQuantity)

	//customers
	app.Post("/customers", controllers.CreateCustomer)
	app.Get("/customers", controllers.GetCustomers)
	app.Get("/customers/:UseID", controllers.GetCustomerByID)
	app.Put("/customers/:customer_id", controllers.UpdateCustomer)

	//hero-slider
	app.Get("/hero-slider", popular.Popular)
	app.Post("/hero-slider", popular.AddPopular)
	app.Put("/hero-slider/:slider_id", popular.UpdatePopular)

}
