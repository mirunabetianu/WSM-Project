package main

import (
	"github.com/gofiber/fiber"
)

func main() {
	// Fiber instance
	app := fiber.New()
	connect()
	// Routes
	app.Post("/payment/pay/:user_id/:order_id/:amount", pay)
	app.Post("/payment/create_user", createUser)
	app.Post("/payment/add_funds/:user_id/:amount", addFunds)
	app.Post("/payment/cancel/:user_id/:order_id", paymentCancel)
	app.Get("/payment/find/:user_id", findUser)
	app.Get("/payment/status/:user_id/:order_id", paymentStatus)

	// start server
	app.Listen(3001)
}

