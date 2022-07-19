package userservice

import "context"

type UserService interface {
	Login(ctx context.Context, email, password string) (User, error)
	Register(ctx context.Context, user User) (User, error)
	FetchUsersByIDs(ctx context.Context, ids []string) (map[string]User, error)
}

type UserRepository interface {
	Register(ctx context.Context, user User) (User, error)
	GetUser(ctx context.Context, identifier string) (User, error)
	FetchUsersByIDs(ctx context.Context, ids []string) (map[string]User, error)
}
