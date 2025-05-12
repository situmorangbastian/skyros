package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orderpb "github.com/situmorangbastian/skyros/proto/order"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	userpb "github.com/situmorangbastian/skyros/proto/user"
	"github.com/situmorangbastian/skyros/serviceutils"
)

func main() {
	log := logrus.New().WithFields(logrus.Fields{"service": "gatewayservice"})
	cfg := viper.New()
	cfg.SetConfigFile(".env")
	cfg.AutomaticEnv()
	err := cfg.ReadInConfig()
	if err != nil {
		log.Fatal("failed read config: ", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(serviceutils.NewRestErrorHandler(log)),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = userpb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, cfg.GetString("USER_SERVICE_GRPC"), opts)
	if err != nil {
		log.Fatalf("Failed to register Service1: %v", err)
	}

	err = productpb.RegisterProductServiceHandlerFromEndpoint(ctx, mux, cfg.GetString("PRODUCT_SERVICE_GRPC"), opts)
	if err != nil {
		log.Fatalf("Failed to register Service2: %v", err)
	}

	err = orderpb.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, cfg.GetString("ORDER_SERVICE_GRPC"), opts)
	if err != nil {
		log.Fatalf("Failed to register Service2: %v", err)
	}

	// Start a single HTTP server for all routes
	log.Info(fmt.Sprintf("Serving gRPC-Gateway on http://localhost:%s", cfg.GetString("PORT")))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.GetInt("PORT")), mux); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
