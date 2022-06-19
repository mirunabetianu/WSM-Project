package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"math"
	utils "order/utils"
	"strconv"
)

var mqtt = utils.OpenMqttConnection()
var database = utils.OpenPsqlConnection()

var stockServiceHost = "localhost"
var stockServicePort = 3001

var paymentServiceHost = "localhost"
var paymentServicePort = 3002

func main() {
	// Fiber instance
	app := fiber.New()

	if utils.GetEnv("STOCK_SERVICE_SERVICE_HOST") != "" {
		stockServiceHost = utils.GetEnv("STOCK_SERVICE_SERVICE_HOST")
	}

	if utils.GetEnv("STOCK_SERVICE_SERVICE_PORT_HTTP") != "" {
		stockServicePort, _ = strconv.Atoi(utils.GetEnv("STOCK_SERVICE_SERVICE_PORT_HTTP"))
	}

	if utils.GetEnv("PAYMENT_SERVICE_SERVICE_HOST") != "" {
		paymentServiceHost = utils.GetEnv("PAYMENT_SERVICE_SERVICE_HOST")
	}

	if utils.GetEnv("PAYMENT_SERVICE_SERVICE_PORT_HTTP") != "" {
		paymentServicePort, _ = strconv.Atoi(utils.GetEnv("PAYMENT_SERVICE_SERVICE_PORT_HTTP"))
	}

	if database == nil {
		fmt.Printf("", database)
		return
	}

	token := mqtt.Subscribe("topic/findItemResponse", 1, nil)
	token.Wait()

	tokenS := mqtt.Subscribe("topic/subtractStockResponse", 1, nil)
	tokenS.Wait()

	tokenP := mqtt.Subscribe("topic/paymentResponse", 1, nil)
	tokenP.Wait()

	tokenR := mqtt.Subscribe("topic/refundResponse", 1, nil)
	tokenR.Wait()

	// Routes
	app.Get("/orders", baseEndpoint)

	// Get all orders
	app.Get("/orders/getAll", getOrders)

	// Get order by order_id
	app.Get("/orders/find/:order_id", findOrder)

	// Create order for user_id
	app.Post("/orders/create/:user_id", createOrder)

	// Remove order by order_id
	app.Delete("/orders/remove/:order_id", removeOrder)

	// Add item to order
	app.Post("/orders/addItem/:order_id/:item_id", addItemToOrder)

	// Remove item from order
	app.Delete("/orders/removeItem/:order_id/:item_id", removeItemFromOrder)

	// Checkout order
	app.Post("/orders/checkout/:order_id", checkout)

	// start server
	err := app.Listen(":3000")
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Handlers
func baseEndpoint(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"status": "running"})
}

func getOrders(c *fiber.Ctx) error {
	var orders []utils.Order

	result := database.Find(&orders)

	if result.Error != nil {
		return c.SendStatus(500)
	}

	return c.Status(200).JSON(orders)
}

func createOrder(c *fiber.Ctx) error {
	order := utils.Order{UserId: c.Params("user_id")}

	result := database.Create(&order)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Could not create order", "data": result.Error.Error()})
	}

	// Return the created order
	return c.Status(200).JSON(fiber.Map{"order_id": order.ID})
}

func removeOrder(c *fiber.Ctx) error {
	id := c.Params("order_id")
	var order utils.Order

	result := database.Delete(&order, id)

	if result.RowsAffected == 0 {
		return c.SendStatus(404)
	}

	return c.SendStatus(200)
}

func findOrder(c *fiber.Ctx) error {
	id := c.Params("order_id")
	var order utils.Order

	result := database.Find(&order, id)

	if result.RowsAffected == 0 {
		return c.SendStatus(404)
	}

	/*
		“order_id”  - the order’s id
		“paid” (true/false)
		“items”  - list of item ids that are included in the order
		“user_id”  - the user’s id that made the order
		“total_cost” - the total cost of the items in the order
	*/

	return c.Status(200).JSON(fiber.Map{"order_id": order.ID, "paid": order.Paid, "items": order.Items, "user_id": order.UserId, "total_cost": order.TotalCost})
}

func addItemToOrder(c *fiber.Ctx) error {
	order_id := c.Params("order_id")
	item_id := c.Params("item_id")

	var order utils.Order

	itemId, errConversionI := strconv.Atoi(item_id)
	orderId, errConversionO := strconv.Atoi(order_id)
	if errConversionI != nil || errConversionO != nil {
		return c.SendStatus(400)
	}

	id := uuid.New()

	channelKey := fmt.Sprintf("orderId:%d-itemId:%d-id:%s", orderId, itemId, id.String())
	token := mqtt.Publish("topic/findItem", 1, false, channelKey)
	token.Wait()

	utils.ItemChannels = append(utils.ItemChannels, utils.ItemChannel{Id: id.String(), OrderId: orderId, ItemId: itemId, Channel: make(chan int)})
	var index int
	for i := range utils.ItemChannels {
		x := len(utils.ItemChannels) - i - 1
		if utils.ItemChannels[x].Id == id.String() {
			index = x
			break
		}
	}

	var itemPrice int
	go func(chan int) {
		itemPrice = <-utils.ItemChannels[index].Channel
	}(utils.ItemChannels[index].Channel)

	if itemPrice == math.MaxInt {
		return c.SendStatus(400)
	} else {
		result := database.Find(&order, order_id)

		if result.RowsAffected == 1 {
			result2 := database.Find(&order, order_id).Updates(utils.Order{Model: order.Model, Paid: order.Paid, UserId: order.UserId, TotalCost: order.TotalCost + itemPrice, Items: append(order.Items, int64(itemId))})

			if result2.RowsAffected == 0 {
				return c.SendStatus(400)
			} else {
				return c.SendStatus(200)
			}
		} else {
			return c.SendStatus(400)
		}
	}
}

func removeItemFromOrder(c *fiber.Ctx) error {
	order_id := c.Params("order_id")
	item_id := c.Params("item_id")

	var order utils.Order

	result := database.Find(&order, order_id)

	if order.Items == nil {
		return c.SendStatus(400)
	}
	if result.RowsAffected == 1 {
		itemId, errConversionI := strconv.Atoi(item_id)
		orderId, errConversionO := strconv.Atoi(order_id)
		if errConversionI != nil || errConversionO != nil {
			return c.SendStatus(400)
		}

		id := uuid.New()

		channelKey := fmt.Sprintf("orderId:%d-itemId:%d-id:%s", orderId, itemId, id.String())
		token := mqtt.Publish("topic/findItem", 1, false, channelKey)
		token.Wait()

		utils.ItemChannels = append(utils.ItemChannels, utils.ItemChannel{Id: id.String(), OrderId: orderId, ItemId: itemId, Channel: make(chan int)})
		var index int
		for i := range utils.ItemChannels {
			x := len(utils.ItemChannels) - i - 1
			if utils.ItemChannels[x].Id == id.String() {
				index = x
				break
			}
		}
		var itemPrice int
		go func(chan int) {
			itemPrice = <-utils.ItemChannels[index].Channel
		}(utils.ItemChannels[index].Channel)
		if itemPrice == math.MaxInt {
			return c.SendStatus(400)
		} else {
			var exist bool
			exist = false
			var items []int64

			for _, s := range order.Items {
				if s == int64(itemId) && exist == false {
					exist = true
				} else {
					items = append(items, s)
				}
			}
			if !exist {
				return c.SendStatus(400)
			}

			result2 := database.Model(&order).Select("TotalCost", "Items").Updates(utils.Order{TotalCost: order.TotalCost - itemPrice, Items: items})

			if result2.RowsAffected == 0 {
				return c.SendStatus(400)
			} else {
				return c.SendStatus(200)
			}
		}

	} else {
		return c.SendStatus(400)
	}
}

func checkout(c *fiber.Ctx) error {
	order_id := c.Params("order_id")

	// find order by id
	var order utils.Order
	result := database.Find(&order, order_id)

	id := uuid.New()

	orderId, errConversionO := strconv.Atoi(order_id)
	items, err := json.Marshal(map[string][]int64{"items": append(order.Items, int64(orderId)), "id": {int64((id.ID()))}})
	if order.Items == nil || err != nil || errConversionO != nil {
		return c.SendStatus(400)
	}
	if result.RowsAffected == 1 {
		channelKey := fmt.Sprintf("orderId:%d-amount:%d-id:%d-userId:%s", orderId, order.TotalCost, id.ID(), order.UserId)
		token := mqtt.Publish("topic/payment", 1, false, channelKey)
		token.Wait()

		tokenN := mqtt.Publish("topic/subtractStock", 1, false, items)
		tokenN.Wait()
		utils.CheckoutChannels = append(utils.CheckoutChannels, utils.CheckoutItem{Id: id.ID(), OrderId: orderId, PaymentChannel: make(chan string), StockChannel: make(chan string)})
		var index int
		for i := range utils.CheckoutChannels {
			x := len(utils.CheckoutChannels) - i - 1
			if utils.CheckoutChannels[x].Id == id.ID() {
				index = x
				break
			}
		}
		var resultPayment, resultStock string
		
		go func(chan string) {
			resultPayment = <-utils.CheckoutChannels[index].PaymentChannel
		}(utils.CheckoutChannels[index].PaymentChannel)

		if resultPayment == "error" {
			return c.SendStatus(404)
		}
		go func(chan string) {
			resultStock = <-utils.CheckoutChannels[index].StockChannel
		}(utils.CheckoutChannels[index].StockChannel)

		if resultStock == "error" || resultPayment == "error" {
			return c.SendStatus(404)
		} else {
			resultUpdateOrder := database.Find(&order, orderId).Update("Paid", true)

			if resultUpdateOrder.Error == nil {
				return c.SendStatus(200)
			}
		}
	}

	return c.SendStatus(404)
}
