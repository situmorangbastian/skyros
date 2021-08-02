package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/situmorangbastian/skyros"
	"github.com/situmorangbastian/skyros/internal"
	handler "github.com/situmorangbastian/skyros/internal/http"
	mysqlRepo "github.com/situmorangbastian/skyros/internal/mysql"
	"github.com/situmorangbastian/skyros/order"
	"github.com/situmorangbastian/skyros/product"
	"github.com/situmorangbastian/skyros/user"
)

func main() {
	// Init Mysql Connection
	dbHost := skyros.GetEnv("MYSQL_HOST")
	dbPort := skyros.GetEnv("MYSQL_PORT")
	dbUser := skyros.GetEnv("MYSQL_USER")
	dbPass := skyros.GetEnv("MYSQL_PASS")
	dbName := skyros.GetEnv("MYSQL_NAME")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
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

	// Migration
	driver, err := mysql.WithInstance(dbConn, &mysql.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/mysql/migrations",
		"mysql", driver)
	if err != nil {
		log.Fatal(err)
	}
	_ = m.Up()

	// Init User
	userRepo := mysqlRepo.NewUserRepository(dbConn)
	userService := user.NewService(userRepo)

	// Init Product
	productRepo := mysqlRepo.NewProductRepository(dbConn)
	productService := product.NewService(productRepo)

	// Init Order
	orderRepo := mysqlRepo.NewOrderRepository(dbConn)
	orderService := order.NewService(orderRepo, productRepo)

	tokenSecretKey := skyros.GetEnv("SECRET_KEY")

	e := echo.New()
	e.Use(
		middleware.JWTWithConfig(middleware.JWTConfig{
			Skipper: func(c echo.Context) bool {
				methodWithPath := c.Request().Method + " " + c.Request().URL.Path
				return handler.AllowEndpointWithoutAuth[methodWithPath]
			},
			SigningKey: tokenSecretKey,
		}),
		handler.ErrorMiddleware(),
	)
	e.Validator = internal.NewValidator()

	// Init Handler
	handler.NewUserHandler(e, userService, tokenSecretKey)
	handler.NewProductHandler(e, productService)
	handler.NewOrderHandler(e, orderService)

	// Start server
	serverAddress := skyros.GetEnv("SERVER_ADDRESS")
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
