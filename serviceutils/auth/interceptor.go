package auth

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

func AuthInterceptor(secretKey string, userClient UserClient) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
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
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, status.Errorf(codes.Unauthenticated, "unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secretKey), nil
			})
			if err != nil || !token.Valid {
				return nil, status.Error(codes.Unauthenticated, "invalid token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, status.Error(codes.Unauthenticated, "invalid token claims")
			}

			userID, ok := claims["id"].(string)
			if !ok || userID == "" {
				return nil, status.Error(codes.Unauthenticated, "invalid token claims")
			}

			users, err := userClient.FetchByIDs(ctx, []string{userID})
			if err != nil {
				return nil, err
			}

			user, exists := users[userID]
			if !exists || user.Email == "" {
				return nil, status.Error(codes.Unauthenticated, "invalid token")
			}

			ctx = context.WithValue(ctx, userClaimsKey, Claims{
				ID:      userID,
				Email:   user.Email,
				Name:    user.Name,
				Address: user.Address,
				Type:    UserType(user.Type),
			})
		}

		return handler(ctx, req)
	}
}

func GetUserClaims(ctx context.Context) (*Claims, error) {
	claims, ok := ctx.Value(userClaimsKey).(Claims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid user")
	}
	return &claims, nil
}

func isGRPCGatewayRequest(md metadata.MD) bool {
	return len(md.Get("grpcgateway-user-agent")) > 0
}
