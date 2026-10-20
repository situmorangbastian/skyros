package auth

import (
	"context"

	userpb "github.com/situmorangbastian/skyros/proto/user"
)

type UserType string

const (
	UserSellerType UserType = "seller"
	UserBuyerType  UserType = "buyer"
)

// Claims holds only what other services need to know about an authenticated user.
// It is NOT the full user domain model — password and internal data are never included.
type Claims struct {
	ID      string   `json:"id"`
	Email   string   `json:"email"`
	Name    string   `json:"name"`
	Address string   `json:"address"`
	Type    UserType `json:"type"`
}

type UserClient interface {
	FetchByIDs(ctx context.Context, ids []string) (map[string]Claims, error)
}

func ToAuthClaims(u *userpb.User) Claims {
	return Claims{
		ID:      u.GetId(),
		Email:   u.GetEmail(),
		Name:    u.GetName(),
		Address: u.GetAddress(),
		Type:    UserType(u.GetType()),
	}
}

// WithClaimsForTest returns a context carrying claims, as AuthInterceptor
// would. It exists so other packages' tests can exercise authenticated paths.
func WithClaimsForTest(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, userClaimsKey, c)
}
