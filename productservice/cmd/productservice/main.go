package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/situmorangbastian/eclipse"
	"github.com/situmorangbastian/skyros/productservice"
	"github.com/situmorangbastian/skyros/productservice/internal"
	grpcHandler "github.com/situmorangbastian/skyros/productservice/internal/grpc"
	handler "github.com/situmorangbastian/skyros/productservice/internal/http"
	mysqlRepo "github.com/situmorangbastian/skyros/productservice/internal/mysql"
	"github.com/situmorangbastian/skyros/productservice/product"
	"github.com/situmorangbastian/skyros/skyrosgrpc"
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

	// Grpc Client
	userServiceGrpcConn, err := grpc.Dial(productservice.GetEnv("USER_SERVICE_GRPC"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	userServiceGrpc := grpcHandler.NewUserService(userServiceGrpcConn)

	// Init Product
	productRepo := mysqlRepo.NewProductRepository(dbConn)
	productService := product.NewService(productRepo, userServiceGrpc)

	tokenSecretKey := productservice.GetEnv("SECRET_KEY")

	e := echo.New()
	e.Use(
		eclipse.Error(),
	)
	e.Validator = internal.NewValidator()

	g := e.Group("")
	g.Use(
		echojwt.JWT([]byte(tokenSecretKey)),
		eclipse.Authentication(),
	)

	// Init Handler
	handler.NewProductHandler(e, g, productService)

	// Start server
	wg := sync.WaitGroup{}
	wg.Add(2)
	serverAddress := productservice.GetEnv("SERVER_ADDRESS")
	go func() {
		defer wg.Done()
		if err := e.Start(serverAddress); err != nil {
			e.Logger.Info("shutting down the server...")
		}
	}()

	grpcServer := grpc.NewServer()
	grpcProductService := grpcHandler.NewProductGrpcServer(productService)
	skyrosgrpc.RegisterProductServiceServer(grpcServer, grpcProductService)

	go func() {
		defer wg.Done()
		port, err := strconv.Atoi(productservice.GetEnv("GRPC_SERVER_ADDRESS"))
		if err != nil {
			log.Fatal(errors.New("invalid grpc server port"))
		}

		listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("GRPC Server Running on Port: ", port)
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}()
	wg.Wait()

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
	grpcServer.GracefulStop()
}
