package productservice

import "context"

type ProductService interface {
	Store(ctx context.Context, product Product) (Product, error)
	Get(ctx context.Context, ID string) (Product, error)
	Fetch(ctx context.Context, filter Filter) ([]Product, string, error)
	FetchByIds(ctx context.Context, ids []string) (map[string]Product, error)
}

type ProductRepository interface {
	Store(ctx context.Context, product Product) (Product, error)
	Get(ctx context.Context, ID string) (Product, error)
	Fetch(ctx context.Context, filter Filter) ([]Product, string, error)
	FetchByIds(ctx context.Context, ids []string) (map[string]Product, error)
}

type UserServiceGrpc interface {
	FetchByIDs(ctx context.Context, ids []string) (map[string]User, error)
}
