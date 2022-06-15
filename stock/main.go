package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	utils "stock/utils"
	"strconv"
)

var database = utils.OpenPsqlConnection()

//var mqtt = utils.OpenMqttConnection()

func main() {
	// Fiber instance
	app := fiber.New()

	// Check utils not null
	if database == nil {
		fmt.Println("Error initializing db and mqtt")
		return
	}

	//utils.Subscribe(mqtt, "topic/addItem")
	//utils.Subscribe(mqtt, "topic/removeItem")

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("hello")
	})

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

	return ctx.Status(200).JSON(&item)
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
