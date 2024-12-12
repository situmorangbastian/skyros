package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	log "github.com/sirupsen/logrus"
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

	// Start server
	go func() {
		if err := app.Listen(address); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Fatal(err)
	}
}
