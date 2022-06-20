package utils

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Credit uint
}

type Payment struct {
	gorm.Model
	Status  byte
	OrderID uint
}

func BaseEndpoint(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"status": "running"})
}

func FindUser(c *fiber.Ctx) error {
	type Item struct {
		ID     uint `json:"user_id"`
		Credit uint `json:"credit"`
	}
	user_id := c.Params("user_id")
	var user User
	result := Database.First(&user, user_id)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(404).JSON(fiber.Map{"error": error_message})
	}
	return c.Status(200).JSON(fiber.Map{"user_id": user.ID, "credit": user.Credit})
}

func CreateUser(c *fiber.Ctx) error {
	type Item struct {
		User_id uint `json:"user_id"`
	}
	user := &User{Credit: 0}
	result := Database.Create(user)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(500).JSON(fiber.Map{"error": error_message})

	}
	return c.Status(200).JSON(fiber.Map{"user_id": user.ID})
}

func AddFunds(c *fiber.Ctx) error {

	var user_id string
	var amount uint
	user_id = c.Params("user_id")
	amount_temp, err := strconv.ParseUint(c.Params("amount"), 10, 64)
	amount = uint(amount_temp)

	if err != nil {
		error_message := fmt.Sprint(err)
		return c.Status(500).JSON(fiber.Map{"done": false, "error": error_message})
	}

	var user User
	result := Database.First(&user, user_id)
	fmt.Printf("result: %v, rows affected %v\n", result.Error, result.RowsAffected)
	if result.Error != nil {
		error_message := fmt.Sprint(result.Error)
		return c.Status(404).JSON(fiber.Map{"done": false, "error": error_message})
	}
	user.Credit = user.Credit + amount
	save_result := Database.Save(&user)
	fmt.Printf("result: %v, rows affected %v\n", save_result.Error, save_result.RowsAffected)
	if save_result.Error != nil || save_result.RowsAffected != 1 {
		error_message := fmt.Sprint(save_result.Error)
		return c.Status(500).JSON(fiber.Map{"done": false, "error": error_message})
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
