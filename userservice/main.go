package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/situmorangbastian/skyros/productservice/middleware"
	userGrpcSvc "github.com/situmorangbastian/skyros/proto/user"
	postgreRepo "github.com/situmorangbastian/skyros/userservice/internal/repository/postgresql"
	"github.com/situmorangbastian/skyros/userservice/internal/service"
	"github.com/situmorangbastian/skyros/userservice/internal/usecase"
	"github.com/situmorangbastian/skyros/userservice/internal/validation"
)

func main() {
	log := logrus.New().WithFields(logrus.Fields{"service": "userservice"})
	cfg := viper.New()
	cfg.SetConfigFile(".env")
	cfg.AutomaticEnv()
	err := cfg.ReadInConfig()
	if err != nil {
		log.Fatal("failed read config: ", err)
	}

	dbConn, err := sql.Open(`postgres`, cfg.GetString("DATABASE_URL"))
	if err != nil {
		log.Fatal("failed database connect: ", err)
	}

	err = dbConn.Ping()
	if err != nil {
		log.Fatal("failed ping database: ", err)
	}
	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal("failed close db connection: ", err)
		}
	}()

	runMigrations(log, cfg.GetString("DATABASE_URL"))

	userRepo := postgreRepo.NewUserRepository(dbConn)
	userUsecase := usecase.NewUserUsecase(userRepo)

	grpcServer := grpc.NewServer()
	userService := service.NewUserService(userUsecase, cfg.GetString("SECRET_KEY"), validation.NewValidator())
	userGrpcSvc.RegisterUserServiceServer(grpcServer, userService)

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(middleware.ErrRestHandler(log)),
	)
	err = userGrpcSvc.RegisterUserServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		cfg.GetString("GRPC_SERVICE_ENDPOINT"),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		log.Fatal("failed register gRPC-Gateway handler: ", err)
	}

	restServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GetInt("GRPC_GATEWAY_SERVER_PORT")),
		Handler: mux,
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		listen, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GetInt("GRPC_SERVER_PORT")))
		if err != nil {
			log.Fatal("failed to listen on network: ", err)
		}

		log.Info("gRPC-Server listening on ", fmt.Sprintf(":%d", cfg.GetInt("GRPC_SERVER_PORT")))
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatal("failed run gRPC-Server: ", err)
		}
	}()

	go func() {
		log.Info("gRPC-Gateway server listening on ", fmt.Sprintf(":%d", cfg.GetInt("GRPC_GATEWAY_SERVER_PORT")))
		if err := restServer.ListenAndServe(); err != nil {
			log.Fatal("failed to serve gRPC-Gateway: ", err)
		}
	}()
	wg.Wait()

	// wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info("shutting down servers...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := restServer.Shutdown(ctx); err != nil {
		log.Error("failed shutdown gRPC-Gateway")
	}
	grpcServer.GracefulStop()
}

func runMigrations(log *logrus.Entry, connStr string) {
	m, err := migrate.New(
		"file://migrations",
		connStr,
	)
	if err != nil {
		log.Fatal("failed to create migrate instance: ", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to apply migrations: ", err)
	}

	log.Info("migrations applied successfull")
}
