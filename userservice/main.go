package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/skyrosgrpc"
	grpcHandler "github.com/situmorangbastian/skyros/userservice/api/grpc"
	restHandler "github.com/situmorangbastian/skyros/userservice/api/rest/handlers"
	"github.com/situmorangbastian/skyros/userservice/api/rest/validators"
	customErrors "github.com/situmorangbastian/skyros/userservice/internal/errors"
	mysqlRepo "github.com/situmorangbastian/skyros/userservice/internal/repository/mysql"
	"github.com/situmorangbastian/skyros/userservice/internal/usecase"
)

func main() {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	// Init Mysql Connection
	dbHost := viper.GetString("MYSQL_HOST")
	dbPort := viper.GetString("MYSQL_PORT")
	dbUser := viper.GetString("MYSQL_USER")
	dbPass := viper.GetString("MYSQL_PASS")
	dbName := viper.GetString("MYSQL_DBNAME")
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

	// Init Repository
	userRepo := mysqlRepo.NewUserRepository(dbConn)

	// Init Usecase
	userService := usecase.NewUserUsecase(userRepo)

	tokenSecretKey := viper.GetString("SECRET_KEY")

	e := echo.New()
	e.Use(
		customErrors.Error(),
	)
	e.Validator = validators.NewValidator()

	g := e.Group("")
	g.Use(
		echojwt.JWT([]byte(tokenSecretKey)),
	)

	// Init Handler
	restHandler.NewUserHandler(e, userService, tokenSecretKey)

	// Start server
	wg := sync.WaitGroup{}
	wg.Add(3)
	serverAddress := viper.GetString("SERVER_ADDRESS")
	go func() {
		defer wg.Done()
		if err := e.Start(serverAddress); err != nil {
			e.Logger.Info("shutting down the server...")
		}
	}()

	grpcServer := grpc.NewServer()
	grpcUserService := grpcHandler.NewUserGrpcServer(userService, tokenSecretKey)
	skyrosgrpc.RegisterUserServiceServer(grpcServer, grpcUserService)

	go func() {
		defer wg.Done()
		listen, err := net.Listen("tcp", fmt.Sprintf(":%d", viper.GetInt("GRPC_SERVER_ADDRESS")))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("GRPC Server Running on Port: ", viper.GetInt("GRPC_SERVER_ADDRESS"))
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}()

	mux := runtime.NewServeMux()
	err = skyrosgrpc.RegisterUserServiceHandlerFromEndpoint(context.Background(), mux, fmt.Sprintf(":%d", viper.GetInt("GRPC_SERVER_ADDRESS")), []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		log.Fatalf("Failed to register gRPC-Gateway handler: %v", err)
	}

	// Start HTTP server for gRPC-Gateway
	go func() {
		log.Printf("gRPC-Gateway server listening on %s", fmt.Sprintf(":%d", viper.GetInt("GRPC_GATEWAY_SERVER_ADDRESS")))
		if err := http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("GRPC_GATEWAY_SERVER_ADDRESS")), mux); err != nil {
			log.Fatalf("Failed to serve gRPC-Gateway: %v", err)
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
