package utils

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func FindUser(c *fiber.Ctx) error {
	id := c.Params("user_id")
	var user User

	result := Database.Find(&user, id)

	if result.RowsAffected == 0 {
		return c.SendStatus(404)
	}

	return c.Status(200).JSON(fiber.Map{"user_id": user.ID, "credit": user.Credit})
}

func CreateUser(c *fiber.Ctx) error {
	user := &User{Credit: 0}
	result := Database.Create(user)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(500).JSON(fiber.Map{"error": error_message})

	}
	return c.Status(200).JSON(fiber.Map{"user_id": user.ID})
}

func AddFunds(c *fiber.Ctx) error {
	var user_id, amount string
	user_id = c.Params("user_id")
	amount = c.Params("amount")

	amountToPay, errConversion := strconv.ParseFloat(amount, 64)

	if errConversion != nil {
		fmt.Println("Conversion error")
		fmt.Println(errConversion)
		return c.Status(500).JSON(fiber.Map{"done": false})
	}

	var user User

	result := Database.Find(&user, user_id).Update("Credit", user.Credit+uint(amountToPay))

	if result.RowsAffected == 0 {
		fmt.Println("0 rows affected")
		return c.Status(500).JSON(fiber.Map{"done": false})
	}
	return c.Status(200).JSON(fiber.Map{"done": true})
}

func Pay(c *fiber.Ctx) error {

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

	var user User
	result := Database.First(&user, user_id)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(404).JSON(fiber.Map{"error": error_message})
	}

	if user.Credit < amount {
		return c.Status(500).JSON(fiber.Map{"error": "not enough credit"})
	}

	user.Credit = user.Credit - amount
	save_result := Database.Save(&user)
	if save_result.Error != nil || save_result.RowsAffected != 1 {
		error_message := fmt.Sprint(save_result.Error)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}

	var payment Payment
	exists := Database.Where(Payment{OrderID: order_id}).First(&payment).Error
	fmt.Printf("payment: %v\n", payment)
	fmt.Printf("error: %v\n", exists)
	if exists == nil {
		return c.Status(500).JSON(fiber.Map{"error": "payment already exists"})
	}

	payment = Payment{Status: 0, OrderID: order_id}

	result_payment := Database.Create(&payment)
	if result_payment.Error != nil {
		error_message := fmt.Sprint(result_payment.Error)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}
	return c.SendStatus(200)
}

func PaymentCancel(c *fiber.Ctx) error {
	//user_id := c.Params("user_id")

	temp_orderid, err := strconv.ParseUint(c.Params("order_id"), 10, 64)
	if err != nil {
		error_message := fmt.Sprint(err)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}
	order_id := uint(temp_orderid)

	var payment Payment
	result := Database.Where(Payment{OrderID: order_id}).First(&payment)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(404).JSON(fiber.Map{"error": error_message})
	}

	payment.Status = 1
	save_result := Database.Save(&payment)
	if save_result.Error != nil || save_result.RowsAffected != 1 {
		error_message := fmt.Sprint(save_result.Error)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}
	return c.SendStatus(200)
}

func PaymentStatus(c *fiber.Ctx) error {
	//user_id := c.Params("user_id")

	temp_orderid, err := strconv.ParseUint(c.Params("order_id"), 10, 64)
	if err != nil {
		error_message := fmt.Sprint(err)
		return c.Status(500).JSON(fiber.Map{"error": error_message})
	}
	order_id := uint(temp_orderid)
	var payment Payment
	result := Database.Where(Payment{OrderID: order_id}).First(&payment)
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

func SubtractAmountLocal(mqttC mqtt.Client, msg mqtt.Message) {
	var orderId, totalCost int
	var userId string

	var id uint32

	_, err := fmt.Sscanf(string(msg.Payload()), "orderId:%d-amount:%d-id:%d-userId:%s", &orderId, &totalCost, &id, &userId)

	var user User
	responseUser := Database.Find(&user, userId)

	notEnoughCredit := (int)(user.Credit)-totalCost < 0

	if err != nil || responseUser.Error != nil || notEnoughCredit {
		payload := fmt.Sprintf("orderId:%d-id:%d-%s", orderId, id, "error")
		token := mqttC.Publish("topic/paymentResponse", 1, false, payload)
		token.Wait()
	} else {
		var payment Payment
		resultPayment := Database.Where(Payment{OrderID: uint(orderId)}).Last(&payment)
		responseUpdate := Database.Find(&user, userId).Updates(User{Credit: uint((int)(user.Credit) - totalCost)})

		if responseUpdate.Error != nil || resultPayment.Error == nil {
			payload := fmt.Sprintf("orderId:%d-id:%d-%s", orderId, id, "error")
			token := mqttC.Publish("topic/paymentResponse", 1, false, payload)
			token.Wait()
		} else {
			payment = Payment{Status: 0, OrderID: uint(orderId)}
			resultCreatePayment := Database.Create(&payment)

			if resultCreatePayment.Error != nil {
				payload := fmt.Sprintf("orderId:%d-id:%d-%s", orderId, id, "error")
				token := mqttC.Publish("topic/paymentResponse", 1, false, payload)
				token.Wait()
			} else {
				payload := fmt.Sprintf("orderId:%d-id:%d-%s", orderId, id, "success")
				token := mqttC.Publish("topic/paymentResponse", 1, false, payload)
				token.Wait()
			}
		}

	}
}

func RefundAmountLocal(mqttC mqtt.Client, msg mqtt.Message) {
	var userId string
	var totalCost int
	var id uint32
	_, err := fmt.Sscanf(string(msg.Payload()), "amount:%d-id:%d-userId:%s", &totalCost, &id, &userId)

	if err != nil {
		payload := fmt.Sprintf("id:%d-%s", id, "error")
		token := mqttC.Publish("topic/refundResponse", 1, false, payload)
		token.Wait()
	} else {
		var user User
		result := Database.Find(&user, userId).Update("Credit", user.Credit+uint(totalCost))

		if result.RowsAffected == 0 {
			payload := fmt.Sprintf("id:%d-%s", id, "error")
			token := mqttC.Publish("topic/refundResponse", 1, false, payload)
			token.Wait()
		} else {
			payload := fmt.Sprintf("id:%d-%s", id, "success")
			token := mqttC.Publish("topic/refundResponse", 1, false, payload)
			token.Wait()
		}
	}
}
