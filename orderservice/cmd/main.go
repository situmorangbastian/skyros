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

	resthandlers "github.com/situmorangbastian/skyros/orderservice/api/rest/handlers"
	"github.com/situmorangbastian/skyros/orderservice/api/rest/middleware"
	"github.com/situmorangbastian/skyros/orderservice/api/rest/validators"
	internalErr "github.com/situmorangbastian/skyros/orderservice/internal/error"
	"github.com/situmorangbastian/skyros/orderservice/internal/helpers"
	mysqlRepo "github.com/situmorangbastian/skyros/orderservice/internal/repository/mysql"
	grpcService "github.com/situmorangbastian/skyros/orderservice/internal/services/grpc"
	"github.com/situmorangbastian/skyros/orderservice/internal/usecase"
)

func main() {
	// Init Mysql Connection
	dbHost := helpers.GetEnv("MYSQL_HOST")
	dbPort := helpers.GetEnv("MYSQL_PORT")
	dbUser := helpers.GetEnv("MYSQL_USER")
	dbPass := helpers.GetEnv("MYSQL_PASS")
	dbName := helpers.GetEnv("MYSQL_DBNAME")
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
	userServiceGrpcConn, err := grpc.Dial(helpers.GetEnv("USER_SERVICE_GRPC"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	productServiceGrpcConn, err := grpc.Dial(helpers.GetEnv("PRODUCT_SERVICE_GRPC"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	userServiceGrpc := grpcService.NewUserService(userServiceGrpcConn)
	productServiceGrpc := grpcService.NewProductService(productServiceGrpcConn)

	// Init Order
	orderRepo := mysqlRepo.NewOrderRepository(dbConn)
	orderService := usecase.NewUsecase(orderRepo, userServiceGrpc, productServiceGrpc)

	tokenSecretKey := helpers.GetEnv("SECRET_KEY")

	e := echo.New()
	e.Use(
		internalErr.Error(),
	)
	e.Validator = validators.NewValidator()

	g := e.Group("")
	g.Use(
		echojwt.JWT([]byte(tokenSecretKey)),
		middleware.Authentication(),
	)

	// Init Handler
	resthandlers.NewOrderHandler(g, orderService)

	// Start server
	serverAddress := helpers.GetEnv("SERVER_ADDRESS")
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
