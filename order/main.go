package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"net/http"
	utils "order/utils"
	"os"
	"strconv"
)

//var mqtt = utils.OpenMqttConnection()
var database = utils.OpenPsqlConnection()

var stockServiceHost = "localhost"
var stockServicePort = 3001

var paymentServiceHost = "localhost"
var paymentServicePort = 3002

func main() {
	// Fiber instance
	app := fiber.New()

	// Routes
	app.Get("/", hello)

	//utils.Subscribe(mqtt, utils.TOPIC_ADD_ITEM)
	//utils.Subscribe(mqtt, utils.TOPIC_REMOVE_ITEM)
	//utils.Publish(mqtt, utils.TOPIC_ADD_ITEM)
	//utils.Publish(mqtt, utils.TOPIC_REMOVE_ITEM)
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
	app.Listen(3000)
}

// Handlers
func hello(c *fiber.Ctx) error {
	return c.SendString("Hello, Order Service!")
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
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Could not create order", "data": result.Error})
	}

	// Return the created order
	return c.Status(200).JSON(fiber.Map{"orderId": order.ID})
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

	return c.Status(200).JSON(&order)
}

func addItemToOrder(c *fiber.Ctx) error {
	orderId := c.Params("order_id")
	itemId := c.Params("item_id")

	var order utils.Order

	//mqtt.Publish(utils.TOPIC_ADD_ITEM, 1, false, itemId)
	item, errConversion := strconv.Atoi(itemId)

	if errConversion != nil {
		return c.SendStatus(400)
	}

	requestURL := fmt.Sprintf("http://%s:%d/stock/find/%d", stockServiceHost, stockServicePort, item)
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}

	//var requestedItem utils.Item

	if res.Status == "500" {
		return c.SendStatus(400)
	} else {
		body, _ := ioutil.ReadAll(res.Body)

		s := string(body)
		requestedItem := utils.Item{}
		err := json.Unmarshal([]byte(s), &requestedItem)
		if err != nil {
			return c.SendStatus(400)
		}

		result := database.Find(&order, orderId)

		if result.RowsAffected == 1 {
			result2 := database.Find(&order, orderId).Updates(utils.Order{Model: order.Model, Paid: order.Paid, UserId: order.UserId, TotalCost: order.TotalCost + int(requestedItem.Price), Items: append(order.Items, int64(item))})

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
	orderId := c.Params("order_id")
	itemId := c.Params("item_id")

	var order utils.Order

	result := database.Find(&order, orderId)

	if order.Items == nil {
		return c.SendStatus(400)
	}
	if result.RowsAffected == 1 {
		item, errConversion := strconv.Atoi(itemId)

		if errConversion != nil {
			return c.SendStatus(400)
		}

		requestURL := fmt.Sprintf("http://%s:%d/stock/find/%d", stockServiceHost, stockServicePort, item)
		res, err := http.Get(requestURL)
		if err != nil {
			fmt.Printf("error making http request: %s\n", err)
			os.Exit(1)
		}

		if res.Status == "500" {
			return c.SendStatus(400)
		} else {
			body, _ := ioutil.ReadAll(res.Body)

			s := string(body)
			requestedItem := utils.Item{}
			err := json.Unmarshal([]byte(s), &requestedItem)
			if err != nil {
				return c.SendStatus(400)
			}

			//var requestedItem utils.Item

			var exist bool
			exist = false

			for i, s := range order.Items {
				if s == int64(item) {
					order.Items[i] = order.Items[len(order.Items)-1]
					exist = true
				}
			}
			if !exist {
				return c.SendStatus(400)
			}

			result2 := database.Find(&order, orderId).Updates(utils.Order{order.Model, order.Paid, order.UserId, order.TotalCost - int(requestedItem.Price), order.Items[:len(order.Items)-1]})

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

//TODO: needs additional endpoints to be implemented
func checkout(c *fiber.Ctx) error {
	orderId := c.Params("order_id")

	// find order by id
	var order utils.Order
	result := database.Find(&order, orderId)

	if order.Items == nil {
		return c.SendStatus(400)
	}
	if result.RowsAffected == 1 {
		// payment to  /payment/pay/{user_id}/{order_id}/{amount}
		requestURL := fmt.Sprintf("http://%s:%d/payment/pay/%d/%d/%d",
			paymentServiceHost,
			paymentServicePort,
			order.UserId,
			orderId,
			order.TotalCost)

		resPaymentService, err := http.Get(requestURL)
		if err != nil {
			fmt.Printf("error making http request: %s\n", err)
			os.Exit(1)
		}

		if resPaymentService.Status == "200" {
			// subtract from /payment/pay/{user_id}/{order_id}/{amount}
			requestURL := fmt.Sprintf("http://%s:%d/stock/subtract/%d/", stockServiceHost, stockServicePort, order.TotalCost)
			res, err := http.Get(requestURL)
			if err != nil {
				fmt.Printf("error making http request: %s\n", err)
				os.Exit(1)
			}
			if res.Status == "400" {
				return c.SendStatus(400)
			}

		} else {
			return c.SendStatus(400)
		}

	}

	return c.SendStatus(500)
}
