package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math"
	"net/http"
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
	app.Post("/orders/checkout/:order_id", checkoutV2)

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

	return c.Status(200).JSON(&order)
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

	channelKey := fmt.Sprintf("orderId:%d-itemId:%d", orderId, itemId)

	token := mqtt.Publish("topic/findItem", 1, false, channelKey)
	token.Wait()

	utils.Chans = append(utils.Chans, utils.ItemChannel{OrderId: orderId, ItemId: itemId, Channel: make(chan int)})
	index := len(utils.Chans) - 1

	itemPrice := <-utils.Chans[index].Channel

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

		channelKey := fmt.Sprintf("orderId:%d-itemId:%d", orderId, itemId)

		token := mqtt.Publish("topic/findItem", 1, false, channelKey)
		token.Wait()

		utils.Chans = append(utils.Chans, utils.ItemChannel{OrderId: orderId, ItemId: itemId, Channel: make(chan int)})
		index := len(utils.Chans) - 1

		itemPrice := <-utils.Chans[index].Channel

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

			result2 := database.Find(&order, orderId).Updates(utils.Order{Model: order.Model, Paid: order.Paid, UserId: order.UserId, TotalCost: order.TotalCost - itemPrice, Items: items})

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

//TODO
// Note: currently we make the subtract call of the stock service for each item we have,
// this might be the bottleneck of the application if we checkout a lot of items,
// I am making the assumption that most orders have few items
func checkout(c *fiber.Ctx) error {
	orderId := c.Params("order_id")

	// find order by id
	var order utils.Order
	result := database.Find(&order, orderId)

	if order.Items == nil {
		return c.SendStatus(400)
	}
	if result.RowsAffected == 1 {
		//emptyPostBody, _ := json.Marshal(map[string]string{})

		// payment to  /payment/pay/{user_id}/{order_id}/{amount}
		paymentRequestUrl := fmt.Sprintf("http://%s:%d/payment/pay/%s/%s/%d",
			paymentServiceHost,
			paymentServicePort,
			order.UserId,
			orderId,
			order.TotalCost)

		fmt.Println(paymentRequestUrl)

		resPaymentService, err := http.Post(paymentRequestUrl, "application/json", nil)

		if err != nil {
			fmt.Printf("error making http request: %s\n", err)
		}

		// keep a list of the item_id for which subtract stock call was successful
		// in case we need to rollback the transaction we need to add the stock again
		var processedItems []int64

		fmt.Println(resPaymentService)
		fmt.Println(resPaymentService.Status)
		fmt.Println(resPaymentService.StatusCode)

		// Subtract stock of all the items via stock service
		if resPaymentService.StatusCode == 200 {
			fmt.Println(order.Items)
			// iterate through all the items of the current order
			for i, s := range order.Items {
				fmt.Println(i, s)
				// TODO we simply subtract 1 for each item id, if we have item_id 5 times, we subtract 1 5 times instead of subtracting once by amount 5
				stockRequestUrl := fmt.Sprintf("http://%s:%d/stock/subtract/%d/1/", stockServiceHost, stockServicePort, s)
				fmt.Println(stockRequestUrl)
				resStockService, err := http.Post(stockRequestUrl, "application/json", nil)

				if err != nil {
					// TODO maybe have a retry with exponential backoff,
					//  sometimes network errors happen, we should have at least a few retries
					//  https://brandur.org/fragments/go-http-retry for reference
					fmt.Printf("error making http request: %s\n", err)
				}

				if resStockService.StatusCode == 200 {
					processedItems = append(processedItems, s)
				} else if resStockService.StatusCode == 400 {
					fmt.Println("Could not subtract stock")
					//rollbackCheckout(order, processedItems)
					return c.SendStatus(400)
				}
			}

		} else {
			fmt.Println("Could not make the payment")
			// return error, payment failed, nothing to rollback
			return c.SendStatus(400)
		}

		// Update the order value in the orders db
		resultUpdateOrder := database.Find(&order, orderId).Update("Paid", true)
		if resultUpdateOrder.RowsAffected == 0 {
			// orders table could not be updated, rollback transaction
			rollbackCheckout(order, processedItems)
			return c.SendStatus(400)
		} else {
			// finally transaction is successful
			return c.SendStatus(200)
		}
	}

	// order not found
	return c.SendStatus(404)
}

func checkoutV2(c *fiber.Ctx) error {
	orderId := c.Params("order_id")

	// find order by id
	var order utils.Order
	result := database.Find(&order, orderId)

	if order.Items == nil {
		return c.SendStatus(400)
	}
	if result.RowsAffected == 1 {
		//emptyPostBody, _ := json.Marshal(map[string]string{})

		// payment to  /payment/pay/{user_id}/{order_id}/{amount}
		paymentRequestUrl := fmt.Sprintf("http://%s:%d/payment/pay/%s/%s/%d",
			paymentServiceHost,
			paymentServicePort,
			order.UserId,
			orderId,
			order.TotalCost)

		fmt.Println(paymentRequestUrl)

		resPaymentService, err := http.Post(paymentRequestUrl, "application/json", nil)

		if err != nil {
			fmt.Printf("error making http request: %s\n", err)
		}

		fmt.Println(resPaymentService)
		fmt.Println(resPaymentService.Status)
		fmt.Println(resPaymentService.StatusCode)

		// Subtract stock of all the items via stock service
		if resPaymentService.StatusCode == 200 {

			// add the array here
			arrayPostBody, _ := json.Marshal(map[string][]int64{"items": order.Items})

			stockRequestUrl := fmt.Sprintf("http://%s:%d/stock/subtract/all/", stockServiceHost, stockServicePort)
			fmt.Println(stockRequestUrl)
			resStockService, err := http.Post(stockRequestUrl, "application/json", bytes.NewBuffer(arrayPostBody))

			if err != nil {
				// TODO maybe have a retry with exponential backoff,
				//  sometimes network errors happen, we should have at least a few retries
				//  https://brandur.org/fragments/go-http-retry for reference
				fmt.Printf("error making http request: %s\n", err)
			}

			if resStockService.StatusCode == 400 {
				fmt.Println("Could not subtract stock")
				return c.SendStatus(400)
			}

		} else {
			fmt.Println("Could not make the payment")
			// return error, payment failed, nothing to rollback
			return c.SendStatus(400)
		}

		// Update the order value in the orders db
		resultUpdateOrder := database.Find(&order, orderId).Update("Paid", true)
		if resultUpdateOrder.RowsAffected == 0 {
			// orders table could not be updated, rollback transaction
			return c.SendStatus(400)
		} else {
			// finally transaction is successful
			return c.SendStatus(200)
		}
	}

	// order not found
	return c.SendStatus(404)
}

// TODO - still need to implement this
func rollbackCheckout(utils.Order, []int64) {
	// cancel the payment that was made
	print()
	// add back the items to the stock that were currently added
}
