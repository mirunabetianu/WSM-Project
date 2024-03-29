package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm/clause"
	utils "stock/utils"
	"strconv"
)

var database = utils.OpenPsqlConnection()
var mqttC = utils.OpenMqttConnection()

func main() {
	// Fiber instance
	app := fiber.New()

	// Check utils not null
	if database == nil {
		fmt.Println("Error initializing db and mqtt")
		return
	}

	token := mqttC.Subscribe("topic/findItem", 1, FindItemLocal)
	token.Wait()

	tokenC := mqttC.Subscribe("topic/subtractStock", 1, SubtractStockLocal)
	tokenC.Wait()
	fmt.Printf("Subscribed to topic: %s", "topic/addItem")

	fmt.Printf("Trying to publish to: %s", "topic/addItem")
	token = mqttC.Publish("topic/addItem", 1, false, "orderId:1-itemId:1")
	token.Wait()

	app.Get("/stock", baseEndpoint)

	// Create new item with given price
	app.Post("/stock/item/create/:price", createItem)

	// Get the stock amount and price of item given id
	app.Get("/stock/find/:item_id", getItem)

	// Subtract stock amount from item given id and amount
	app.Post("/stock/subtract/:item_id/:amount", subtractStockFromItem)

	// Subtract stock amount from the array of items, happening during order checkout
	app.Post("/stock/subtract/all", subtractStockFromItems)

	// Add stock amount to the item
	app.Post("/stock/add/:item_id/:amount", addStockToItem)

	// Start the server
	err := app.Listen(":3001")
	if err != nil {
		return
	}
}

func baseEndpoint(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"status": "running"})
}

func createItem(ctx *fiber.Ctx) error {
	price, _ := strconv.Atoi(ctx.Params("price"))
	if price < 0 {
		ctx.Status(400)
	}

	item := utils.Item{Price: uint(price)}

	result := database.Create(&item)

	if result.Error != nil {
		return ctx.Status(500).JSON(fiber.Map{"status": "error", "message": "Could not create item", "data": result.Error})
	} else {
		// Return the item_id of the created item
		return ctx.Status(200).JSON(fiber.Map{"item_id": item.ID})
	}
}

func getItem(ctx *fiber.Ctx) error {
	item_id := ctx.Params("item_id")
	var item utils.Item

	result := database.Find(&item, item_id)

	if result.RowsAffected == 0 {
		return ctx.SendStatus(404)
	}

	return ctx.Status(200).JSON(fiber.Map{"stock": item.Stock, "price": item.Price})
}

func subtractStockFromItem(ctx *fiber.Ctx) error {
	item_id := ctx.Params("item_id")
	amount, _ := strconv.Atoi(ctx.Params("amount"))

	if amount < 0 {
		ctx.Status(400)
	}

	var item utils.Item

	result := database.Find(&item, item_id)

	if result.RowsAffected == 1 && item.Stock >= uint(amount) {
		result2 := database.Find(&item, item_id).Update("Stock", item.Stock-uint(amount))

		if result2.RowsAffected == 0 {
			return ctx.SendStatus(400)
		} else {
			return ctx.SendStatus(200)
		}
	} else {
		return ctx.SendStatus(400)
	}
}

func subtractStockFromItems(ctx *fiber.Ctx) error {
	var body map[string][]int64

	err := json.Unmarshal(ctx.Body(), &body)

	if err != nil {
		return ctx.SendStatus(400)
	}

	fmt.Println(body)

	return ctx.SendStatus(200)
}

func addStockToItem(ctx *fiber.Ctx) error {
	item_id := ctx.Params("item_id")
	amount, _ := strconv.Atoi(ctx.Params("amount"))

	if amount < 0 {
		ctx.Status(400)
	}

	var item utils.Item

	result := database.Find(&item, item_id).Update("Stock", item.Stock+uint(amount))

	if result.RowsAffected == 0 {
		return ctx.SendStatus(404)
	} else {
		return ctx.SendStatus(200)
	}
}

func FindItemLocal(client mqtt.Client, msg mqtt.Message) {
	var orderId, itemId int

	var id string

	_, err := fmt.Sscanf(string(msg.Payload()), "orderId:%d-itemId:%d-id:%s", &orderId, &itemId, &id)

	var item utils.Item
	result := database.Find(&item, itemId)

	var status int
	if result.RowsAffected == 0 || err != nil {
		status = 500
	} else {
		status = 200
	}

	finalResult := fmt.Sprintf("orderId:%d-itemId:%d-price:%d-status:%d-id:%s", orderId, itemId, item.Price, status, id)

	token := mqttC.Publish("topic/findItemResponse", 1, false, finalResult)
	token.Wait()
}

func SubtractStockLocal(client mqtt.Client, msg mqtt.Message) {
	var body map[string][]int64

	err := json.Unmarshal(msg.Payload(), &body)

	itemIds := body["items"]
	id := uint32(body["id"][0])

	orderId := itemIds[len(itemIds)-1]

	itemIds = itemIds[:len(itemIds)-1]

	dict := make(map[int64]uint)
	for _, num := range itemIds {
		dict[num] = dict[num] + 1
	}

	var notEnoughStock bool
	notEnoughStock = false

	var itemRows []utils.Item
	database.Table("items").Where("id IN ?", itemIds).Select("id", "stock", "price").Scan(&itemRows)

	itemRowsCopy := make([]utils.Item, len(itemRows))
	copy(itemRowsCopy, itemRows)

	for index, targetRow := range itemRows {
		if targetRow.Stock >= dict[int64(targetRow.ID)] {
			itemRows[index].Stock = targetRow.Stock - dict[int64(targetRow.ID)]
		} else {
			notEnoughStock = true
		}
	}

	//for index, value := range dict {
	//	var item utils.Item
	//	resultItem := database.Find(&item, index)
	//
	//	if resultItem.Error != nil || int(item.Stock)-int(value) < 0 {
	//		notEnoughStock = true
	//	}
	//}

	if err != nil || notEnoughStock {
		payload := fmt.Sprintf("orderId:%d-id:%d-%s", orderId, id, "error")
		token := mqttC.Publish("topic/subtractStockResponse", 1, false, payload)
		token.Wait()
	} else {
		result := database.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"stock"}),
		}).Create(&itemRows)

		//var anyError bool
		//anyError = false
		//for index, value := range dict {
		//	var item utils.Item
		//
		//	resultItem := database.Find(&item, index).Update("Stock", (uint)(int(item.Stock)-int(value)))
		//
		//	if resultItem.Error != nil {
		//		anyError = true
		//	}
		//}

		if int(result.RowsAffected) < len(dict) {
			payload := fmt.Sprintf("orderId:%d-id:%d-%s", orderId, id, "error")
			token := mqttC.Publish("topic/subtractStockResponse", 1, false, payload)
			token.Wait()

			database.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"stock"}),
			}).Create(&itemRowsCopy)

		} else {
			payload := fmt.Sprintf("orderId:%d-id:%d-%s", orderId, id, "success")
			token := mqttC.Publish("topic/subtractStockResponse", 1, false, payload)
			token.Wait()
		}
	}

}
