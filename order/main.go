package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	Paid      bool   `gorm:"type:bool;default:false"`
	UserId    string `gorm:"type:varchar;not null"`
	TotalCost int    `gorm:"type:bigint;default:0"`
	Items     []Item `gorm:"many2many:order_item;"`
}

// Item TODO
type Item struct {
	gorm.Model
	Stock  int     `gorm:"type:bigint;default:0"`
	Price  int     `gorm:"type:bigint;default:0"`
	Orders []Order `gorm:"many2many:order_item;"`
}

var database = openPsqlConnection()

func main() {
	// Fiber instance
	app := fiber.New()

	if database == nil {
		fmt.Printf("", database)
		return
	}

	// Routes
	app.Get("/", hello)
	app.Get("/orders/getAll", getOrders)
	app.Get("/orders/find/:order_id", findOrder)

	// Endpoint: /orders/create/{user_id}
	// Method POST - creates an order for the given user, and returns an order_id
	// Output JSON fields: “order_id”  - the order’s id
	app.Post("/orders/create/:user_id", createOrder)

	app.Delete("/orders/remove/:order_id", removeOrder)

	app.Post("/orders/addItem/:order_id/:item_id", addItemToOrder)

	app.Delete("/orders/addItem/:order_id/:item_id", removeItemFromOrder)

	app.Post("/orders/checkout/:order_id", checkout)

	// start server
	err := app.Listen(":3000")
	if err != nil {
		return
	}
}

// Handlers
func hello(c *fiber.Ctx) error {
	return c.SendString("Hello, Order Service!")
}

func getOrders(c *fiber.Ctx) error {
	var orders []Order

	result := database.Find(&orders)

	if result.Error != nil {
		return c.SendStatus(500)
	}

	return c.Status(200).JSON(orders)
}

func createOrder(c *fiber.Ctx) error {
	order := Order{UserId: c.Params("user_id")}

	result := database.Create(&order)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Could not create order", "data": result.Error})
	}

	// Return the created order
	return c.Status(200).JSON(fiber.Map{"orderId": order.ID})
}

func removeOrder(c *fiber.Ctx) error {
	id := c.Params("order_id")
	var order Order

	result := database.Delete(&order, id)

	if result.RowsAffected == 0 {
		return c.SendStatus(404)
	}

	return c.SendStatus(200)
}

func findOrder(c *fiber.Ctx) error {
	id := c.Params("order_id")
	var order Order

	result := database.Find(&order, id)

	if result.RowsAffected == 0 {
		return c.SendStatus(404)
	}

	return c.Status(200).JSON(&order)
}

//TODO
func addItemToOrder(c *fiber.Ctx) error {
	return c.SendStatus(500)
}

//TODO
func removeItemFromOrder(c *fiber.Ctx) error {
	return c.SendStatus(500)
}

//TODO: needs additional endpoints to be implemented
func checkout(c *fiber.Ctx) error {
	return c.SendStatus(500)
}
