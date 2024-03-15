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
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/situmorangbastian/skyros/orderservice"
	"github.com/situmorangbastian/skyros/orderservice/internal"
	grpcHandler "github.com/situmorangbastian/skyros/orderservice/internal/grpc"
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

	// Grpc Client
	userServiceGrpcConn, err := grpc.Dial(orderservice.GetEnv("USER_SERVICE_GRPC"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	productServiceGrpcConn, err := grpc.Dial(orderservice.GetEnv("PRODUCT_SERVICE_GRPC"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	userServiceGrpc := grpcHandler.NewUserService(userServiceGrpcConn)
	productServiceGrpc := grpcHandler.NewProductService(productServiceGrpcConn)

	// Init Order
	orderRepo := mysqlRepo.NewOrderRepository(dbConn)
	orderService := order.NewService(orderRepo, userServiceGrpc, productServiceGrpc)

	tokenSecretKey := orderservice.GetEnv("SECRET_KEY")

	e := echo.New()
	e.Use(
		orderservice.Error(),
	)
	e.Validator = internal.NewValidator()

	g := e.Group("")
	g.Use(
		echojwt.JWT([]byte(tokenSecretKey)),
		orderservice.Authentication(),
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
