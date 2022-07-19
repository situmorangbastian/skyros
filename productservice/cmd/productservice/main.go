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

	"github.com/situmorangbastian/skyros/productservice"
	"github.com/situmorangbastian/skyros/productservice/internal"
	handler "github.com/situmorangbastian/skyros/productservice/internal/http"
	mysqlRepo "github.com/situmorangbastian/skyros/productservice/internal/mysql"
	"github.com/situmorangbastian/skyros/productservice/product"
)

func main() {
	// Init Mysql Connection
	dbHost := productservice.GetEnv("MYSQL_HOST")
	dbPort := productservice.GetEnv("MYSQL_PORT")
	dbUser := productservice.GetEnv("MYSQL_USER")
	dbPass := productservice.GetEnv("MYSQL_PASS")
	dbName := productservice.GetEnv("MYSQL_DBNAME")
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

	// Init Product
	productRepo := mysqlRepo.NewProductRepository(dbConn)
	productService := product.NewService(productRepo)

	tokenSecretKey := productservice.GetEnv("SECRET_KEY")

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
	handler.NewProductHandler(e, g, productService)

	// Start server
	serverAddress := productservice.GetEnv("SERVER_ADDRESS")
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
