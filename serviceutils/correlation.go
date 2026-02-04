package serviceutils

import (
	"context"

	"github.com/kenshaw/stringid"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	errpb "github.com/situmorangbastian/skyros/proto/errors"
)

const (
	CorrelationIDKey = "x-correlation-id"
)

type contextKey string

const correlationIDContextKey = contextKey(CorrelationIDKey)

func extractOrGenerateCorrelationID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if ids := md.Get(CorrelationIDKey); len(ids) > 0 {
			return ids[0]
		}
	}
	return stringid.Generate()
}

func setCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDContextKey, correlationID)
}

func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDContextKey).(string); ok {
		return id
	}
	return ""
}

func UnaryServerInterceptorWithLogging() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		correlationID := extractOrGenerateCorrelationID(ctx)

		ctx = setCorrelationID(ctx, correlationID)

		logger := log.With().
			Str("correlation_id", correlationID).
			Str("method", info.FullMethod).
			Logger()
		ctx = logger.WithContext(ctx)

		logger.Info().Msg("gRPC request received")

		resp, err := handler(ctx, req)

		if err != nil {
			logger.Error().Err(err).Msg("gRPC request failed")
		} else {
			logger.Info().Msg("gRPC request completed")
		}

		return resp, err
	}
}

func TraceErrors() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {

			st, ok := status.FromError(err)
			switch {
			case !ok:
				return resp, err
			default:

				// if st.Details is already populated, we don't want to overwrite
				if len(st.Details()) > 0 {
					return resp, err
				}
				newErr := WithTraceID(ctx, st)
				return resp, newErr
			}
		}

		return resp, err
	}
}

func WithTraceID(ctx context.Context, st *status.Status) error {
	newSt, err := st.WithDetails(&errpb.Errors{
		TraceId: GetCorrelationID(ctx),
	})
	if err != nil {
		return st.Err()
	}
	return newSt.Err()
}
