package main

import (
	"github.com/gofiber/fiber/v2"
	"payment/utils"
)

var mqttC = utils.OpenMqttConnection()

func main() {
	// Fiber instance
	app := fiber.New()
	err, _ := utils.OpenPsqlConnection()
	if err != nil {
		return
	}

	token := mqttC.Subscribe("topic/payment", 1, utils.SubtractAmountLocal)
	token.Wait()

	// Routes
	//app.Get("/payment", utils.BaseEndpoint)
	app.Post("/payment/pay/:user_id/:order_id/:amount", utils.Pay)
	app.Post("/payment/create_user", utils.CreateUser)
	app.Post("/payment/add_funds/:user_id/:amount", utils.AddFunds)
	app.Post("/payment/cancel/:user_id/:order_id", utils.PaymentCancel)
	app.Get("/payment/find_user/:user_id", utils.FindUser)
	app.Get("/payment/status/:user_id/:order_id", utils.PaymentStatus)

	// start server
	app.Listen(":3002")
}
