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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpcIntg "github.com/situmorangbastian/skyros/productservice/internal/integration/grpc"
	"github.com/situmorangbastian/skyros/productservice/internal/repository/postgresql"
	"github.com/situmorangbastian/skyros/productservice/internal/service"
	"github.com/situmorangbastian/skyros/productservice/internal/usecase"
	"github.com/situmorangbastian/skyros/productservice/internal/validation"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	"github.com/situmorangbastian/skyros/serviceutils"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "orderservice").
		Caller().
		Logger()

	cfg := viper.New()
	cfg.SetConfigFile(".env")
	if err := cfg.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal().Err(err).Msg("failed read config")
		}
	}
	cfg.AutomaticEnv()

	if cfg.GetString("APP_ENV") == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: zerolog.TimeFormatUnix,
		}
		log.Logger = zerolog.New(output).
			With().
			Timestamp().
			Str("service", "orderservice").
			Caller().
			Logger()
	}

	dbConn, err := sql.Open(`postgres`, cfg.GetString("DATABASE_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed connect database")
	}

	err = dbConn.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("failed ping database")
	}
	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("failed close connection database")
		}
	}()

	err = runMigrations(cfg.GetString("DATABASE_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed run migrations")
	}

	log.Info().Msg("run migrations successfully")

	userClient, err := grpc.NewClient(
		cfg.GetString("USER_SERVICE_GRPC"),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msg("failed init userservice client")
	}
	defer userClient.Close()

	usrIntgClient := grpcIntg.NewUserIntegrationClient(userClient)

	productRepo := postgresql.NewProductRepository(dbConn)
	productUsecase := usecase.NewProductUsecase(productRepo, usrIntgClient)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(serviceutils.AuthInterceptor(cfg.GetString("SECRET_KEY"))),
	)
	productService := service.NewProductService(productUsecase, validation.NewValidator())
	productpb.RegisterProductServiceServer(grpcServer, productService)

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(serviceutils.NewRestErrorHandler()),
	)
	err = productpb.RegisterProductServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		cfg.GetString("GRPC_SERVICE_ENDPOINT"),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed register gRPC-Gateway")
	}

	restServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GetInt("GRPC_GATEWAY_SERVER_PORT")),
		Handler: mux,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		listen, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GetInt("GRPC_SERVER_PORT")))
		if err != nil {
			log.Fatal().Err(err).Msg("failed listen on network")
		}

		log.Info().Str("port", cfg.GetString("GRPC_SERVER_PORT")).Msg("gRPC-Server starting")
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatal().Err(err).Msg("failed run gRPC-Server")
		}
	}()

	if cfg.GetBool("ENABLE_GATEWAY_GRPC") {
		wg.Add(1)
		go func() {
			log.Info().Str("port", cfg.GetString("GRPC_GATEWAY_SERVER_PORT")).Msg("gRPC-Gateway server starting")
			if err := restServer.ListenAndServe(); err != nil {
				log.Fatal().Err(err).Msg("failed run gRPC-Gateway server")
			}
		}()
	}
	wg.Wait()

	// wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info().Msg("shutting down servers...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := restServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("failed shutdown gRPC-Gatewat")
	}
	grpcServer.GracefulStop()
}

func runMigrations(connStr string) error {
	m, err := migrate.New(
		"file://migrations",
		connStr,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
