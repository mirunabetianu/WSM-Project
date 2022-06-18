package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"payment/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func findUser(c *fiber.Ctx) error {
	type Item struct {
		ID     uint `json:"user_id"`
		Credit uint `json:"credit"`
	}
	user_id := c.Params("user_id")
	var user utils.User
	result := utils.Database.First(&user, user_id)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(404).JSON(fiber.Map{"error": error_message})
	}
	return c.Status(200).JSON(fiber.Map{"user_id": user.ID, "credit": user.Credit})
}

func createUser(c *fiber.Ctx) error {
	type Item struct {
		User_id uint `json:"user_id"`
	}
	user := &utils.User{Credit: 0}
	result := utils.Database.Create(user)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(500).JSON(fiber.Map{"error": error_message})

	}
	return c.Status(200).JSON(fiber.Map{"user_id": user.ID})
}

func addFunds(c *fiber.Ctx) error {

	var user_id string
	var amount uint
	user_id = c.Params("user_id")
	amount_temp, err := strconv.ParseUint(c.Params("amount"), 10, 64)
	amount = uint(amount_temp)

	if err != nil {
		error_message := fmt.Sprint(err)
		return c.Status(500).JSON(fiber.Map{"done": false, "error": error_message})
	}

	var user utils.User
	result := utils.Database.First(&user, user_id)
	fmt.Printf("result: %v, rows affected %v\n", result.Error, result.RowsAffected)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(404).JSON(fiber.Map{"done": false, "error": error_message})
	}
	user.Credit = user.Credit + amount
	save_result := utils.Database.Save(&user)
	fmt.Printf("result: %v, rows affected %v\n", save_result.Error, save_result.RowsAffected)
	if save_result.Error != nil || save_result.RowsAffected != 1 {
		error_message := fmt.Sprint(save_result.Error)
		return c.Status(500).JSON(fiber.Map{"done": false, "error": error_message})
	}
	return c.Status(200).JSON(fiber.Map{"done": true})
}

func pay(c *fiber.Ctx) error {

	user_id := c.Params("user_id")
	temp_orderid, err := strconv.ParseUint(c.Params("order_id"), 10, 64)
	if err != nil {
		error_message := fmt.Sprint(err)
		return c.Status(404).JSON(fiber.Map{"error": error_message})
	}
	order_id := uint(temp_orderid)

	temp_amount, err := strconv.ParseUint(c.Params("amount"), 10, 64)
	if err != nil {
		error_message := fmt.Sprint(err)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}
	amount := uint(temp_amount)

	var user utils.User
	result := utils.Database.First(&user, user_id)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(404).JSON(fiber.Map{"error": error_message})
	}

	if user.Credit < amount {
		return c.Status(500).JSON(fiber.Map{"error": "not enough credit"})
	}

	user.Credit = user.Credit - amount
	save_result := utils.Database.Save(&user)
	if save_result.Error != nil || save_result.RowsAffected != 1 {
		error_message := fmt.Sprint(save_result.Error)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}

	var payment utils.Payment
	exists := utils.Database.Where(utils.Payment{OrderID: order_id}).First(&payment).Error
	fmt.Printf("payment: %v\n", payment)
	fmt.Printf("error: %v\n", exists)
	if exists == nil {
		return c.Status(500).JSON(fiber.Map{"error": "payment already exists"})
	}

	payment = utils.Payment{Status: 0, OrderID: order_id}

	result_payment := utils.Database.Create(&payment)
	if result_payment.Error != nil {
		error_message := fmt.Sprint(result_payment.Error)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}
	return c.SendStatus(200)
}

func paymentCancel(c *fiber.Ctx) error {
	//user_id := c.Params("user_id")

	temp_orderid, err := strconv.ParseUint(c.Params("order_id"), 10, 64)
	if err != nil {
		error_message := fmt.Sprint(err)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}
	order_id := uint(temp_orderid)

	var payment utils.Payment
	result := utils.Database.Where(utils.Payment{OrderID: order_id}).First(&payment)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(404).JSON(fiber.Map{"error": error_message})
	}

	payment.Status = 1
	save_result := utils.Database.Save(&payment)
	if save_result.Error != nil || save_result.RowsAffected != 1 {
		error_message := fmt.Sprint(save_result.Error)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}
	return c.SendStatus(200)
}

func paymentStatus(c *fiber.Ctx) error {
	//user_id := c.Params("user_id")

	temp_orderid, err := strconv.ParseUint(c.Params("order_id"), 10, 64)
	if err != nil {
		error_message := fmt.Sprint(err)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}
	order_id := uint(temp_orderid)
	var payment utils.Payment
	result := utils.Database.Where(utils.Payment{OrderID: order_id}).First(&payment)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(404).JSON(fiber.Map{"error": error_message})
	}
	var paid bool
	if payment.Status == 0 {
		paid = true
	} else {
		paid = false
	}

	return c.Status(200).JSON(fiber.Map{"paid": paid})
}

func SubtractAmountLocal(client mqtt.Client, msg mqtt.Message) {
	var orderId, totalCost int
	var userId string

	print(string(msg.Payload()))
	_, err := fmt.Sscanf(string(msg.Payload()), "orderId:%d-amount:%d-userId:%s", &orderId, &totalCost, &userId)

	var user utils.User
	responseUser := utils.Database.Find(&user, userId)

	notEnoughCredit := user.Credit-(uint(totalCost)) < 0

	println(err != nil)
	println(responseUser.Error != nil)
	println(notEnoughCredit)
	if err != nil || responseUser.Error != nil || notEnoughCredit {
		payload := fmt.Sprintf("orderId:%d-%s", orderId, "error")
		token := mqttC.Publish("topic/paymentResponse", 1, false, payload)
		token.Wait()
	} else {
		var payment utils.Payment
		resultPayment := utils.Database.Where(utils.Payment{OrderID: uint(orderId)}).First(&payment)
		responseUpdate := utils.Database.Find(&user, userId).Updates(utils.User{Credit: user.Credit - (uint(totalCost))})

		if responseUpdate.Error != nil || resultPayment.Error == nil {
			payload := fmt.Sprintf("orderId:%d-%s", orderId, "error")
			token := mqttC.Publish("topic/paymentResponse", 1, false, payload)
			token.Wait()
		} else {
			payment = utils.Payment{Status: 0, OrderID: uint(orderId)}
			resultCreatePayment := utils.Database.Create(&payment)

			if resultCreatePayment.Error != nil {
				payload := fmt.Sprintf("orderId:%d-%s", orderId, "error")
				token := mqttC.Publish("topic/paymentResponse", 1, false, payload)
				token.Wait()
			} else {
				payload := fmt.Sprintf("orderId:%d-%s", orderId, "success")
				token := mqttC.Publish("topic/paymentResponse", 1, false, payload)
				token.Wait()
			}
		}

	}
}
