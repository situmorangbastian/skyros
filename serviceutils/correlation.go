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

func setCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDContextKey, correlationID)
}

func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDContextKey).(string); ok {
		return id
	}
	return ""
}

func CorrelationServerInterceptorWithLogging() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		var correlationID string
		if values := md.Get(CorrelationIDKey); len(values) > 0 {
			correlationID = values[0]
		}
		if correlationID == "" {
			correlationID = stringid.Generate()
		}

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

func CorrelationClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		corrID := GetCorrelationID(ctx)
		if corrID != "" {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			}
			md = md.Copy()
			md.Set(CorrelationIDKey, corrID)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
