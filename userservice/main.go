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

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	cfg "github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/skyrosgrpc"
	grpcHandler "github.com/situmorangbastian/skyros/userservice/api/grpc"
	mysqlRepo "github.com/situmorangbastian/skyros/userservice/internal/repository/mysql"
	"github.com/situmorangbastian/skyros/userservice/internal/usecase"
)

func main() {
	log := logrus.New().WithFields(logrus.Fields{"service": "userservice"})
	cfg.SetConfigFile(".env")
	cfg.AutomaticEnv()
	err := cfg.ReadInConfig()
	if err != nil {
		log.Fatal("failed read config: ", err)
	}

	dbConn, err := sql.Open(`mysql`, cfg.GetString("DATABASE_URL"))
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

	userRepo := mysqlRepo.NewUserRepository(dbConn)
	userService := usecase.NewUserUsecase(userRepo)

	grpcServer := grpc.NewServer()
	grpcUserService := grpcHandler.NewUserGrpcServer(userService, cfg.GetString("SECRET_KEY"))
	skyrosgrpc.RegisterUserServiceServer(grpcServer, grpcUserService)

	mux := runtime.NewServeMux()
	err = skyrosgrpc.RegisterUserServiceHandlerFromEndpoint(context.Background(), mux, fmt.Sprintf(":%d", cfg.GetInt("GRPC_SERVER_PORT")), []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		log.Fatal("failed register gRPC-Gateway handler: ", err)
	}

	grpcGatewaySvr := &http.Server{
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
		if err := grpcGatewaySvr.ListenAndServe(); err != nil {
			log.Fatal("Failed to serve gRPC-Gateway: ", err)
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
	if err := grpcGatewaySvr.Shutdown(ctx); err != nil {
		log.Error("failed shutdown gRPC-Gateway")
	}
	grpcServer.GracefulStop()
}
