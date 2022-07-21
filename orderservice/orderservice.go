package orderservice

import "context"

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

type UserServiceGrpc interface {
	FetchByIDs(ctx context.Context, ids []string) (map[string]User, error)
}

type ProductServiceGrpc interface {
	FetchByIDs(ctx context.Context, ids []string) (map[string]Product, error)
}
