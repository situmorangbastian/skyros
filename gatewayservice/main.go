package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Endpoint []Endpoint
}

type Endpoint struct {
	Path    string
	Service string
}

var (
	config Config
)

func init() {
	configFile := "config.toml"

	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err.Error())
	}

	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatalln(err.Error())
	}
}

func main() {
	fiberApp := fiber.New()

	for _, endpoint := range config.Endpoint {
		fiberApp.All(endpoint.Path, func(c *fiber.Ctx) error {
			target := fmt.Sprintf("%s/%s", endpoint.Service, endpoint.Path)
			return proxy.Do(c, target)
		})
	}

	address := viper.GetString("server.address")
	if address == "" {
		log.Fatalln("address is not set")
	}

	// Start server
	go func() {
		if err := fiberApp.Listen(address); err != nil {
			log.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	if err := fiberApp.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Fatal(err)
	}
}
