package serviceutils

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const userClaimsKey contextKey = "userClaims"

func AuthInterceptor(secretKey string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		if isGRPCGatewayRequest(md) {
			authHeaders := md.Get("authorization")
			if len(authHeaders) == 0 || !strings.HasPrefix(authHeaders[0], "Bearer ") {
				return nil, status.Error(codes.Unauthenticated, "missing or invalid bearer token")
			}

			tokenStr := strings.TrimPrefix(authHeaders[0], "Bearer ")
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})
			if err != nil || !token.Valid {
				return nil, status.Error(codes.Unauthenticated, "invalid token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, status.Error(codes.Unauthenticated, "invalid token claims")
			}

			ctx = context.WithValue(ctx, userClaimsKey, claims)
		}
		return handler(ctx, req)
	}
}

func GetUserClaims(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(jwt.MapClaims)
	return claims, ok
}

func isGRPCGatewayRequest(md metadata.MD) bool {
	// grpc-gateway adds this header automatically
	if vals := md.Get("grpcgateway-user-agent"); len(vals) > 0 {
		return true
	}
	return false
}
