package main

import "github.com/gofiber/fiber"

func main() {
	// Fiber instance
	app := fiber.New()

	// Routes
	app.Get("/", hello)

	// start server
	app.Listen(3001)
}

// Handler
func hello(c *fiber.Ctx) {
	c.Send("Hello, Payment Service!")
}
