package main

import (
	"context"
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
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	grpcClient "github.com/situmorangbastian/skyros/orderservice/internal/integration/grpc"
	"github.com/situmorangbastian/skyros/orderservice/internal/repository/postgresql"
	"github.com/situmorangbastian/skyros/orderservice/internal/service"
	"github.com/situmorangbastian/skyros/orderservice/internal/usecase"
	"github.com/situmorangbastian/skyros/orderservice/internal/validation"
	orderpb "github.com/situmorangbastian/skyros/proto/order"
	"github.com/situmorangbastian/skyros/serviceutils"
	"github.com/situmorangbastian/skyros/serviceutils/auth"
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
			log.Error().Err(err).Msg("failed read config")
			log.Info().Msg("configure using automatic env")
			cfg.AutomaticEnv()
		}
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbpool, err := pgxpool.New(ctx, cfg.GetString("DATABASE_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed connect database")
	}
	defer dbpool.Close()

	err = dbpool.Ping(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed ping database")
	}

	err = runMigrations(cfg.GetString("DATABASE_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed run migrations")
	}

	log.Info().Msg("run migrations successfully")

	userSvcClient, err := grpc.NewClient(cfg.GetString("USER_SERVICE_GRPC"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msg("failed init userservice client")
	}
	productSvcClient, err := grpc.NewClient(cfg.GetString("PRODUCT_SERVICE_GRPC"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msg("failed init productservice client")
	}
	userClient := grpcClient.NewUserClient(userSvcClient)
	productClient := grpcClient.NewProductClient(productSvcClient)

	orderRepo := postgresql.NewOrderRepository(dbpool)
	orderUsecase := usecase.NewUsecase(orderRepo, userClient, productClient, log.Logger)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(auth.AuthInterceptor(cfg.GetString("SECRET_KEY"), userClient)),
	)
	orderService := service.NewOrderService(orderUsecase, validation.NewValidator(), log.Logger)
	orderpb.RegisterOrderServiceServer(grpcServer, orderService)

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
		}),
		runtime.WithErrorHandler(serviceutils.NewRestErrorHandler()),
	)
	err = orderpb.RegisterOrderServiceHandlerFromEndpoint(
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
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
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
