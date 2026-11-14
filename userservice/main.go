package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
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
	os.Exit(run())
}

// run starts the service and returns the process exit code. It returns rather
// than exiting so that deferred cleanup runs when a server fails at runtime.
func run() int {
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
		if errors.Is(err, fs.ErrNotExist) {
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

	if err := runMigrations(cfg.GetString("DATABASE_URL"), cfg.GetString("MIGRATIONS_DIR")); err != nil {
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
		runtime.WithMetadata(serviceutils.GatewayMetadataAnnotator),
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
		Addr:              fmt.Sprintf(":%d", cfg.GetInt("GRPC_GATEWAY_SERVER_PORT")),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	var gateway *http.Server
	if cfg.GetBool("ENABLE_GATEWAY_GRPC") {
		gateway = restServer
	}

	if err := serviceutils.Serve(grpcServer, fmt.Sprintf(":%d", cfg.GetInt("GRPC_SERVER_PORT")), gateway); err != nil {
		return 1
	}
	return 0
}

func runMigrations(connStr, dir string) error {
	if dir == "" {
		dir = "migrations"
	}
	m, err := migrate.New("file://"+dir, connStr)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
