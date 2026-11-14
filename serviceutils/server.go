package serviceutils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// ShutdownTimeout bounds how long in-flight HTTP requests get to finish.
const ShutdownTimeout = 10 * time.Second

// Serve runs grpcServer on grpcAddr and restServer, either of which may be
// nil, until the process receives SIGINT/SIGTERM or one of the servers fails.
// Both are then shut down gracefully.
//
// It returns the server error that caused the shutdown, or nil after a clean
// signal-driven stop. Callers return instead of calling log.Fatal, so their
// deferred cleanup (DB pools, client connections) still runs.
func Serve(grpcServer *grpc.Server, grpcAddr string, restServer *http.Server) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return serve(ctx, grpcServer, grpcAddr, restServer)
}

func serve(ctx context.Context, grpcServer *grpc.Server, grpcAddr string, restServer *http.Server) error {
	var listener net.Listener
	if grpcServer != nil {
		var err error
		listener, err = net.Listen("tcp", grpcAddr)
		if err != nil {
			return fmt.Errorf("listen on %s: %w", grpcAddr, err)
		}
	}

	// buffered so a failing server never blocks after we stop listening
	serverErr := make(chan error, 2)

	if grpcServer != nil {
		go func() {
			log.Info().Str("addr", listener.Addr().String()).Msg("gRPC server starting")
			if err := grpcServer.Serve(listener); err != nil {
				serverErr <- fmt.Errorf("gRPC server: %w", err)
			}
		}()
	}

	if restServer != nil {
		go func() {
			log.Info().Str("addr", restServer.Addr).Msg("HTTP gateway starting")
			if err := restServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				serverErr <- fmt.Errorf("HTTP gateway: %w", err)
			}
		}()
	}

	var runErr error
	select {
	case <-ctx.Done():
		log.Info().Msg("shutdown signal received")
	case runErr = <-serverErr:
		log.Error().Err(runErr).Msg("server failed, shutting down")
	}

	if restServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
		defer cancel()
		if err := restServer.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("HTTP gateway shutdown")
		}
	}
	if grpcServer != nil {
		grpcServer.GracefulStop()
	}

	log.Info().Msg("servers stopped")
	return runErr
}
