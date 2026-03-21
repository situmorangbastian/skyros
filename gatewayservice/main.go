package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderpb "github.com/situmorangbastian/skyros/proto/order"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	userpb "github.com/situmorangbastian/skyros/proto/user"
	"github.com/situmorangbastian/skyros/serviceutils"
)

func main() {
	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", "gatewayservice").
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

	required := []string{"PORT", "USER_SERVICE_GRPC", "PRODUCT_SERVICE_GRPC", "ORDER_SERVICE_GRPC"}
	for _, key := range required {
		if cfg.GetString(key) == "" {
			log.Fatal().Str("key", key).Msg("missing required config")
		}
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(serviceutils.NewRestErrorHandler()),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := userpb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, cfg.GetString("USER_SERVICE_GRPC"), opts); err != nil {
		log.Fatal().Err(err).Msg("failed to register user service")
	}

	if err := productpb.RegisterProductServiceHandlerFromEndpoint(ctx, mux, cfg.GetString("PRODUCT_SERVICE_GRPC"), opts); err != nil {
		log.Fatal().Err(err).Msg("failed to register product service")
	}

	if err := orderpb.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, cfg.GetString("ORDER_SERVICE_GRPC"), opts); err != nil {
		log.Fatal().Err(err).Msg("failed to register order service")
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.GetInt("PORT")),
		Handler: mux,
	}

	go func() {
		log.Info().Str("port", cfg.GetString("PORT")).Msg("gRPC-Gateway server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to run gRPC-Gateway server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}
	log.Info().Msg("server exited")
}
