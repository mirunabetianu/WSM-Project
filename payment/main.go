package main

import (
	"github.com/gofiber/fiber/v2"
	utils "payment/utils"
)

func main() {
	// Fiber instance
	app := fiber.New()
	utils.Connect()
	// Routes
	app.Get("/payment", utils.BaseEndpoint)
	app.Post("/payment/pay/:user_id/:order_id/:amount", utils.Pay)
	app.Post("/payment/create_user", utils.CreateUser)
	app.Post("/payment/add_funds/:user_id/:amount", utils.AddFunds)
	app.Post("/payment/cancel/:user_id/:order_id", utils.PaymentCancel)
	app.Get("/payment/find/:user_id", utils.FindUser)
	app.Get("/payment/status/:user_id/:order_id", utils.PaymentStatus)

	// start server
	app.Listen(":3002")
}
