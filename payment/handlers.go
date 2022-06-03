package main

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

func findUser(c *fiber.Ctx) error {
	type Item struct {
		ID     uint `json:"user_id"`
		Credit uint `json:"credit"`
	}
	user_id := c.Params("user_id")
	var user User
	result := Database.First(&user, user_id)
	if result.Error != nil {
		return c.SendStatus(400)

	}
	return c.Status(200).JSON(Item{user.ID, user.Credit})
}

func createUser(c *fiber.Ctx) error {
	type Item struct {
		User_id uint `json:"user_id"`
	}
	user := &User{Credit: 0}
	result := Database.Create(user)
	if result.Error != nil {
		return c.SendStatus(400)

	}
	return c.Status(200).JSON(Item{user.ID})
}

func addFunds(c *fiber.Ctx) error {
	type Item struct {
		Done bool `json:"done"`
	}

	var user_id string
	var amount uint
	user_id = c.Params("user_id")
	amount_temp, err := strconv.ParseUint(c.Params("amount"), 10, 64)
	amount = uint(amount_temp)

	if err != nil {
		return c.Status(400).JSON(Item{false})
	}

	var user User
	result := Database.First(&user, user_id)
	fmt.Printf("result: %v, rows affected %v\n", result.Error, result.RowsAffected)
	if result.Error != nil {
		return c.Status(400).JSON(Item{false})
	}
	user.Credit = user.Credit + amount
	save_result := Database.Save(&user)
	fmt.Printf("result: %v, rows affected %v\n", save_result.Error, save_result.RowsAffected)
	if save_result.Error != nil || save_result.RowsAffected != 1 {
		return c.Status(400).JSON(Item{false})
	}
	return c.Status(200).JSON(Item{true})
}

func pay(c *fiber.Ctx) error {

	user_id := c.Params("user_id")
	temp_orderid, err := strconv.ParseUint(c.Params("order_id"), 10, 64)
	if err != nil {
		c.SendStatus(400)
	}
	order_id := uint(temp_orderid)

	temp_amount, err := strconv.ParseUint(c.Params("amount"), 10, 64)
	if err != nil {
		return c.SendStatus(400)
	}
	amount := uint(temp_amount)

	var user User
	result := Database.First(&user, user_id)
	if result.Error != nil {
		return c.SendStatus(400)
	}

	if user.Credit < amount {
		return c.SendStatus(400)
	}

	user.Credit = user.Credit - amount
	save_result := Database.Save(&user)
	if save_result.Error != nil || save_result.RowsAffected != 1 {
		return c.SendStatus(400)
	}

	var payment Payment
	exists := Database.Where(Payment{OrderID: order_id}).First(&payment).Error
	fmt.Printf("payment: %v\n", payment)
	fmt.Printf("error: %v\n", exists)
	if exists == nil {
		return c.SendStatus(400)
	}

	payment = Payment{Status: 0, OrderID: order_id}

	result_payment := Database.Create(&payment)
	if result_payment.Error != nil {
		return c.SendStatus(400)
	}
	return c.SendStatus(200)
}

func paymentCancel(c *fiber.Ctx) error {

	//user_id := c.Params("user_id")

	temp_orderid, err := strconv.ParseUint(c.Params("order_id"), 10, 64)
	if err != nil {
		return c.SendStatus(400)
	}
	order_id := uint(temp_orderid)

	var payment Payment
	result := Database.Where(Payment{OrderID: order_id}).First(&payment)
	if result.Error != nil {
		return c.SendStatus(400)
	}

	payment.Status = 1
	save_result := Database.Save(&payment)
	if save_result.Error != nil || save_result.RowsAffected != 1 {
		return c.SendStatus(400)
	}
	return c.SendStatus(200)
}

func paymentStatus(c *fiber.Ctx) error {
	//user_id := c.Params("user_id")

	type Order struct {
		Paid bool `json:"done"`
	}
	order := new(Order)
	call("3000", "/orders/find/1/", order)
	return c.Status(200).JSON(order)

	//this logic is in case the payment status is kept in a different table in the payment database.

	// type Item struct {
	// 	Paied bool `json:"paid"`
	// }
	// temp_orderid, err := strconv.ParseUint(c.Params("order_id"), 10, 64)
	// if err != nil {
	// 	return c.SendStatus(400)
	// }
	// order_id := uint(temp_orderid)
	// var payment Payment
	// result := Database.Where(Payment{OrderID: order_id}).First(&payment)
	// if result.Error != nil {
	// 	return c.SendStatus(400)
	// }
	// var paid bool
	// if payment.Status == 0 {
	// 	paid = true
	// } else {
	// 	paid = false
	// }

	// return c.Status(200).JSON(Item{paid})

}
