package orderservice

import "context"

type UserService interface {
	Login(ctx context.Context, email, password string) (User, error)
	Register(ctx context.Context, user User) (User, error)
}

type UserRepository interface {
	Register(ctx context.Context, user User) (User, error)
	GetUser(ctx context.Context, identifier string) (User, error)
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

type OrderService interface {
	Store(ctx context.Context, order Order) (Order, error)
	Get(ctx context.Context, ID string) (Order, error)
	Fetch(ctx context.Context, filter Filter) ([]Order, string, error)
	PatchStatus(ctx context.Context, ID string, status int) error
}

type OrderRepository interface {
	Store(ctx context.Context, order Order) (Order, error)
	Fetch(ctx context.Context, filter Filter) ([]Order, string, error)
	PatchStatus(ctx context.Context, ID string, status int) error
}