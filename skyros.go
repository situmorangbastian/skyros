package skyros

import "context"

type UserService interface {
	Login(ctx context.Context, email, password string) (User, error)
	Register(ctx context.Context, user User) (User, error)
}

type UserRepository interface {
	Register(ctx context.Context, user User) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
}
