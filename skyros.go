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

type ProductService interface {
	Store(ctx context.Context, product Product) (Product, error)
	Get(ctx context.Context, ID string) (Product, error)
	Fetch(ctx context.Context, filter Filter) ([]Product, string, error)
}

type ProductRepository interface {
	Store(ctx context.Context, product Product) (Product, error)
	Get(ctx context.Context, ID string) (Product, error)
	Fetch(ctx context.Context, filter Filter) ([]Product, string, error)
}
