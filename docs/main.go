package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func main() {
	app := fiber.New()

	app.Static("/docs", "./docs.yaml")

	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL: "http://localhost:8080/docs",
	}))

	app.Listen(":8080")
}
