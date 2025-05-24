package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

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
	cfg.SetConfigFile(".env")
	if err := cfg.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Error().Err(err).Msg("failed read config")
			log.Info().Msg("configure using automatic env")
			cfg.AutomaticEnv()
		}
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(serviceutils.NewRestErrorHandler()),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := userpb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, cfg.GetString("USER_SERVICE_GRPC"), opts)
	if err != nil {
		log.Fatal().Err(err).Msg("failed register user service")
	}

	err = productpb.RegisterProductServiceHandlerFromEndpoint(ctx, mux, cfg.GetString("PRODUCT_SERVICE_GRPC"), opts)
	if err != nil {
		log.Fatal().Err(err).Msg("failed register product service")
	}

	err = orderpb.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, cfg.GetString("ORDER_SERVICE_GRPC"), opts)
	if err != nil {
		log.Fatal().Err(err).Msg("failed register order service")
	}

	log.Info().Str("port", cfg.GetString("PORT")).Msg("gRPC-Gateway server starting")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.GetInt("PORT")), mux); err != nil {
		log.Fatal().Err(err).Msg("failed run gRPC-Gateway server")
	}
}
