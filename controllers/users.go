package controllers

import (
	"context"
	"dog/condb"
	"dog/models"
	"dog/utils"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

func GetEmployees(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), `
        SELECT id, employee_id, password, name, address, phone, email, position, salary, hire_date, created_at
        FROM employee
        ORDER BY id ASC
    `)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var employees []models.Employee

	for rows.Next() {
		var emp models.Employee
		if err := rows.Scan(
			&emp.ID,
			&emp.EmployeeID,
			&emp.Password,
			&emp.Name,
			&emp.Address,
			&emp.Phone,
			&emp.Email,
			&emp.Position,
			&emp.Salary,
			&emp.HireDate,
			&emp.CreatedAt,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		employees = append(employees, emp)
	}

	return c.JSON(employees)
}

func GetEmployeeByID(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	employeeID := c.Params("employee_id")

	var emp models.Employee
	err = conn.QueryRow(context.Background(), `
        SELECT id, employee_id, password, name, address, phone, email, position, salary, hire_date, created_at
        FROM employee
        WHERE employee_id = $1
    `, employeeID).Scan(
		&emp.ID,
		&emp.EmployeeID,
		&emp.Password,
		&emp.Name,
		&emp.Address,
		&emp.Phone,
		&emp.Email,
		&emp.Position,
		&emp.Salary,
		&emp.HireDate,
		&emp.CreatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
	}

	return c.JSON(emp)
}

func GenerateNextEmployeeID(conn *pgx.Conn) (string, error) {
	var lastID string

	err := conn.QueryRow(
		context.Background(),
		`SELECT employee_id 
		 FROM employee 
		 WHERE employee_id LIKE 'EMP%' 
		 ORDER BY employee_id DESC 
		 LIMIT 1`,
	).Scan(&lastID)

	if err != nil {
		// ถ้าไม่มีข้อมูลในระบบเลย ให้เริ่ม EMP001
		if err == pgx.ErrNoRows {
			return "EMP001", nil
		}
		return "", err
	}

	// ตัด EMP แล้วแปลงเป็น int
	numStr := strings.TrimPrefix(lastID, "EMP")
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return "", fmt.Errorf("invalid employee_id format: %v", err)
	}

	newID := fmt.Sprintf("EMP%03d", num+1)
	return newID, nil
}

func CreateEmployee(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	newEmpID, err := GenerateNextEmployeeID(conn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// ดึงข้อมูลอื่นจาก body
	var emp models.Employee
	if err := c.BodyParser(&emp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// เพิ่ม employee_id ใหม่
	emp.EmployeeID = newEmpID

	// Insert
	_, err = conn.Exec(context.Background(),
		`INSERT INTO employee (employee_id, password, name, address, phone, email, position, salary, hire_date)
     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_DATE)`,
		emp.EmployeeID,
		emp.Password,
		emp.Name,
		emp.Address,
		emp.Phone,
		emp.Email,
		emp.Position,
		emp.Salary,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Employee created",
		"employee_id": newEmpID,
	})
}

func UpdateEmployee(c *fiber.Ctx) error {
	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB connection failed"})
	}
	defer conn.Close(context.Background())

	employeeID := c.Params("employee_id")
	if employeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employee_id is required"})
	}

	var updateData models.Employee // ใช้ struct ให้ตรงกับตาราง employee
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	commandTag, err := conn.Exec(context.Background(),
		`UPDATE employee SET name=$1, address=$2, phone=$3, email=$4, updated_at=NOW() WHERE employee_id=$5`,
		updateData.Name,
		updateData.Address,
		updateData.Phone,
		updateData.Email,
		employeeID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if commandTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
	}

	return c.JSON(fiber.Map{
		"message": "Employee updated successfully",
	})
}

func Login(c *fiber.Ctx) error {

	var U models.User_input
	if err := c.BodyParser(&U); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input: " + err.Error(),
		})
	}

	conn, err := condb.DB_Lek()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "DB connection failed",
		})
	}
	defer conn.Close(
		context.Background())

	var emp models.Employee
	query := `SELECT employee_id, password, name, position FROM employee WHERE employee_id = $1`
	err = conn.QueryRow(context.Background(), query, U.EmployeeID).Scan(
		&emp.EmployeeID,
		&emp.Password,
		&emp.Name,
		&emp.Position,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Employee not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Query failed: " + err.Error(),
		})
	}

	if emp.Password != U.Password {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Incorrect password",
		})
	}

	//  Bcrypt
	// if err := bcrypt.CompareHashAndPassword([]byte(emp.Password), []byte(U.Password)); err != nil {
	// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
	// 		"error": "Incorrect password",
	// 	})
	// }

	token, err := utils.GenerateJWTToken(emp.EmployeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Token generation failed",
		})
	}

	utils.SetJWTCookie(c, token)

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"employee": fiber.Map{
			"employee_id": emp.EmployeeID,
			"name":        emp.Name,
		},
		"token": token,
	})
}
