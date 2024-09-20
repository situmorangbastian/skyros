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

	"github.com/situmorangbastian/skyros/skyrosgrpc"
	grpcHandler "github.com/situmorangbastian/skyros/userservice/api/grpc"
	restHandler "github.com/situmorangbastian/skyros/userservice/api/rest/handlers"
	"github.com/situmorangbastian/skyros/userservice/api/rest/validators"
	"github.com/situmorangbastian/skyros/userservice/internal/config"
	internalErr "github.com/situmorangbastian/skyros/userservice/internal/errors"
	mysqlRepo "github.com/situmorangbastian/skyros/userservice/internal/repository/mysql"
	"github.com/situmorangbastian/skyros/userservice/internal/usecase"
)

func main() {
	// Init Mysql Connection
	dbHost := config.GetEnv("MYSQL_HOST")
	dbPort := config.GetEnv("MYSQL_PORT")
	dbUser := config.GetEnv("MYSQL_USER")
	dbPass := config.GetEnv("MYSQL_PASS")
	dbName := config.GetEnv("MYSQL_DBNAME")
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

	tokenSecretKey := config.GetEnv("SECRET_KEY")

	e := echo.New()
	e.Use(
		internalErr.Error(),
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
	wg.Add(2)
	serverAddress := config.GetEnv("SERVER_ADDRESS")
	go func() {
		defer wg.Done()
		if err := e.Start(serverAddress); err != nil {
			e.Logger.Info("shutting down the server...")
		}
	}()

	grpcServer := grpc.NewServer()
	grpcUserService := grpcHandler.NewUserGrpcServer(userService)
	skyrosgrpc.RegisterUserServiceServer(grpcServer, grpcUserService)

	go func() {
		defer wg.Done()
		port, err := strconv.Atoi(config.GetEnv("GRPC_SERVER_ADDRESS"))
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
