package main

import (
	"github.com/gofiber/fiber"
	utils "order/utils"
)

var mqtt = utils.OpenMqttConnection()

func main() {
	// Fiber instance
	app := fiber.New()

	// Routes
	app.Get("/", hello)

	utils.Subscribe(mqtt, "topic/wdm")
	utils.Publish(mqtt, "topic/wdm")

	// Endpoint: /orders/create/{user_id}
	// Method POST - creates an order for the given user, and returns an order_id
	// Output JSON fields: “order_id”  - the order’s id
	//app.Post("/orders/create/:user_id", func(c *fiber.Ctx) {
	//	var connectionOpen = openPsqlConnection()
	//	if connectionOpen == "connection open" {
	//		c.SendString("Created order for the given user " + c.Params("user_id"))
	//	} else {
	//		c.SendString("Failed to create order for the given user " + c.Params("user_id"))
	//	}
	//})

	// start server
	app.Listen(3000)
}

// Handlers
func hello(c *fiber.Ctx) {
	c.Send("Hello, Order Service!")
}
