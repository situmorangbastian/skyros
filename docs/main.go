package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/spf13/viper"
)

func init() {
	configFile := "config.toml"

	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}

}

func main() {
	address := viper.GetString("server.address")
	if address == "" {
		log.Fatal("address is not well-set")
	}

	app := fiber.New()

	app.Static("/apidocs", "./docs.yaml")

	app.Get("/docs/*", swagger.New(swagger.Config{
		URL: fmt.Sprintf("http://localhost%s/apidocs", address),
	}))

	app.Listen(address)
}
