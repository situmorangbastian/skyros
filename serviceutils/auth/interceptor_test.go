package auth

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testSecret = "test-secret"

type fakeUserClient struct{ users map[string]Claims }

func (f fakeUserClient) FetchByIDs(_ context.Context, ids []string) (map[string]Claims, error) {
	out := map[string]Claims{}
	for _, id := range ids {
		if c, ok := f.users[id]; ok {
			out[id] = c
		}
	}
	return out, nil
}

func signedToken(t *testing.T, id string) string {
	t.Helper()
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": id}).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func call(t *testing.T, md metadata.MD) (*Claims, error) {
	t.Helper()
	client := fakeUserClient{users: map[string]Claims{
		"u1": {ID: "u1", Email: "a@b.c", Type: UserBuyerType},
	}}
	var got *Claims
	handler := func(ctx context.Context, _ any) (any, error) {
		got, _ = GetUserClaims(ctx)
		return nil, nil
	}
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err := AuthInterceptor(testSecret, client)(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	return got, err
}

func TestAuthInterceptor(t *testing.T) {
	t.Run("internal call without gateway marker passes through unauthenticated", func(t *testing.T) {
		claims, err := call(t, metadata.MD{})
		if err != nil || claims != nil {
			t.Fatalf("claims=%v err=%v, want nil/nil", claims, err)
		}
	})

	t.Run("gateway request without User-Agent still requires a token", func(t *testing.T) {
		_, err := call(t, metadata.Pairs(gatewayMarkerKey, "true"))
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", status.Code(err))
		}
	})

	t.Run("valid token populates claims", func(t *testing.T) {
		claims, err := call(t, metadata.Pairs(
			gatewayMarkerKey, "true",
			"authorization", "Bearer "+signedToken(t, "u1"),
		))
		if err != nil {
			t.Fatal(err)
		}
		if claims == nil || claims.ID != "u1" || claims.Type != UserBuyerType {
			t.Fatalf("unexpected claims: %+v", claims)
		}
	})

	t.Run("token for unknown user is rejected", func(t *testing.T) {
		_, err := call(t, metadata.Pairs(
			gatewayMarkerKey, "true",
			"authorization", "Bearer "+signedToken(t, "ghost"),
		))
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", status.Code(err))
		}
	})

	t.Run("token signed with another key is rejected", func(t *testing.T) {
		bad, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": "u1"}).SignedString([]byte("other"))
		_, err := call(t, metadata.Pairs(gatewayMarkerKey, "true", "authorization", "Bearer "+bad))
		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("code = %v, want Unauthenticated", status.Code(err))
		}
	})
}
