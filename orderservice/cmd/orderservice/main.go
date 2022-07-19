package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/situmorangbastian/skyros/orderservice"
	"github.com/situmorangbastian/skyros/orderservice/internal"
	handler "github.com/situmorangbastian/skyros/orderservice/internal/http"
	mysqlRepo "github.com/situmorangbastian/skyros/orderservice/internal/mysql"
	"github.com/situmorangbastian/skyros/orderservice/order"
)

func main() {
	// Init Mysql Connection
	dbHost := orderservice.GetEnv("MYSQL_HOST")
	dbPort := orderservice.GetEnv("MYSQL_PORT")
	dbUser := orderservice.GetEnv("MYSQL_USER")
	dbPass := orderservice.GetEnv("MYSQL_PASS")
	dbName := orderservice.GetEnv("MYSQL_DBNAME")
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	dbConn, err := sql.Open(`mysql`, dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = dbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Init Order
	orderRepo := mysqlRepo.NewOrderRepository(dbConn)
	orderService := order.NewService(orderRepo)

	tokenSecretKey := orderservice.GetEnv("SECRET_KEY")

	e := echo.New()
	e.Use(
		handler.ErrorMiddleware(),
	)
	e.Validator = internal.NewValidator()

	g := e.Group("")
	g.Use(
		middleware.JWT([]byte(tokenSecretKey)),
		handler.Authentication(),
	)

	// Init Handler
	handler.NewOrderHandler(g, orderService)

	// Start server
	serverAddress := orderservice.GetEnv("SERVER_ADDRESS")
	go func() {
		if err := e.Start(serverAddress); err != nil {
			e.Logger.Info("shutting down the server...")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
