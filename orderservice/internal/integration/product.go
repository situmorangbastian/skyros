package integration

import (
	"context"

	"github.com/situmorangbastian/skyros/orderservice/internal/models"
)

type ProductClient interface {
	FetchByIDs(ctx context.Context, ids []string) (map[string]models.Product, error)
}
