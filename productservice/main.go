package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	grpcIntg "github.com/situmorangbastian/skyros/productservice/internal/integration/grpc"
	"github.com/situmorangbastian/skyros/productservice/internal/repository/postgresql"
	"github.com/situmorangbastian/skyros/productservice/internal/service"
	"github.com/situmorangbastian/skyros/productservice/internal/usecase"
	"github.com/situmorangbastian/skyros/productservice/internal/validation"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	"github.com/situmorangbastian/skyros/serviceutils"
	"github.com/situmorangbastian/skyros/serviceutils/auth"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "productservice").
		Caller().
		Logger()

	cfg := viper.New()
	cfg.AutomaticEnv()
	cfg.SetConfigFile(".env")

	if err := cfg.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info().Msg(".env not found, using environment variables")
		} else {
			log.Fatal().Err(err).Msg("failed to read config file")
		}
	}

	if cfg.GetString("APP_ENV") == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: zerolog.TimeFormatUnix,
		}).With().Timestamp().Str("service", "productservice").Caller().Logger()
	}

	required := []string{
		"DATABASE_URL",
		"SECRET_KEY",
		"GRPC_SERVER_PORT",
		"GRPC_SERVICE_ENDPOINT",
		"USER_SERVICE_GRPC",
	}

	for _, key := range required {
		if cfg.GetString(key) == "" {
			log.Fatal().Str("key", key).Msg("missing required config")
		}
	}

	dbpool, err := pgxpool.New(context.Background(), cfg.GetString("DATABASE_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer dbpool.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbpool.Ping(pingCtx); err != nil {
		log.Fatal().Err(err).Msg("failed to ping database")
	}

	if err := runMigrations(cfg.GetString("DATABASE_URL")); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("migrations applied successfully")

	userSvcConn, err := grpc.NewClient(
		cfg.GetString("USER_SERVICE_GRPC"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(serviceutils.CorrelationClientInterceptor()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init user service client")
	}

	usrIntgClient := grpcIntg.NewUserIntegrationClient(userSvcConn)

	productRepo := postgresql.NewProductRepository(dbpool)
	productUsecase := usecase.NewProductUsecase(productRepo, usrIntgClient)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			serviceutils.CorrelationServerInterceptorWithLogging(),
			serviceutils.TraceErrors(),
			auth.AuthInterceptor(cfg.GetString("SECRET_KEY"), usrIntgClient),
		),
	)

	productService := service.NewProductService(productUsecase, validation.NewValidator())
	productpb.RegisterProductServiceServer(grpcServer, productService)

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithErrorHandler(serviceutils.NewRestErrorHandler()),
	)

	if err := productpb.RegisterProductServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		cfg.GetString("GRPC_SERVICE_ENDPOINT"),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	); err != nil {
		log.Fatal().Err(err).Msg("failed to register gRPC-Gateway")
	}

	restServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GetInt("GRPC_GATEWAY_SERVER_PORT")),
		Handler: mux,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GetInt("GRPC_SERVER_PORT")))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to listen on network")
		}

		log.Info().Str("port", cfg.GetString("GRPC_SERVER_PORT")).Msg("gRPC server starting")

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal().Err(err).Msg("gRPC server failed")
		}
	}()

	if cfg.GetBool("ENABLE_GATEWAY_GRPC") {
		wg.Add(1)
		go func() {
			defer wg.Done()

			log.Info().Str("port", cfg.GetString("GRPC_GATEWAY_SERVER_PORT")).Msg("gRPC-Gateway starting")

			if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("gRPC-Gateway failed")
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down servers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := restServer.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown gRPC-Gateway")
	}

	grpcServer.GracefulStop()

	if err := userSvcConn.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close user service gRPC connection")
	}

	wg.Wait()
	log.Info().Msg("servers exited")
}

func runMigrations(connStr string) error {
	m, err := migrate.New("file://migrations", connStr)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
