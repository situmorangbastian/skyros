package repository

import (
	"context"

	"github.com/situmorangbastian/skyros/productservice/internal/models"
)

type ProductRepository interface {
	Store(ctx context.Context, product models.Product) (models.Product, error)
	Get(ctx context.Context, ID string) (models.Product, error)
	Fetch(ctx context.Context, filter models.ProductFilter) ([]models.Product, error)
	FetchByIds(ctx context.Context, ids []string) (map[string]models.Product, error)
}
