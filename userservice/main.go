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

	userpb "github.com/situmorangbastian/skyros/proto/user"
	"github.com/situmorangbastian/skyros/serviceutils"
	"github.com/situmorangbastian/skyros/userservice/internal/repository/postgresql"
	"github.com/situmorangbastian/skyros/userservice/internal/service"
	"github.com/situmorangbastian/skyros/userservice/internal/usecase"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "userservice").
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
		}).With().Timestamp().Str("service", "userservice").Caller().Logger()
	}

	required := []string{"DATABASE_URL", "SECRET_KEY", "GRPC_SERVER_PORT", "GRPC_SERVICE_ENDPOINT"}
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

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := dbpool.Ping(pingCtx); err != nil {
		log.Fatal().Err(err).Msg("failed to ping database")
	}

	if err := runMigrations(cfg.GetString("DATABASE_URL")); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("migrations applied successfully")

	userRepo := postgresql.NewUserRepository(dbpool)
	userUsecase := usecase.NewUserUsecase(userRepo, log.Logger)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			serviceutils.CorrelationServerInterceptorWithLogging(),
			serviceutils.TraceErrors(),
		),
	)
	userService := service.NewUserService(userUsecase, cfg.GetString("SECRET_KEY"), serviceutils.NewCustomValidator(), log.Logger)
	userpb.RegisterUserServiceServer(grpcServer, userService)

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
	if err := userpb.RegisterUserServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		cfg.GetString("GRPC_SERVICE_ENDPOINT"),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	); err != nil {
		log.Fatal().Err(err).Msg("failed to register gRPC-Gateway handler")
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
			log.Fatal().Err(err).Msg("failed to listen on network")
		}
		log.Info().Str("port", cfg.GetString("GRPC_SERVER_PORT")).Msg("gRPC server starting")
		if err := grpcServer.Serve(listen); err != nil {
			log.Fatal().Err(err).Msg("failed to run gRPC server")
		}
	}()

	if cfg.GetBool("ENABLE_GATEWAY_GRPC") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Info().Str("port", cfg.GetString("GRPC_GATEWAY_SERVER_PORT")).Msg("gRPC-Gateway server starting")
			if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("failed to run gRPC-Gateway server")
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
