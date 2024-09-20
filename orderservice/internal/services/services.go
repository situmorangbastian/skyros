package services

import (
	"context"

	"github.com/situmorangbastian/skyros/orderservice/internal/domain/models"
)

type UserServiceGrpc interface {
	FetchByIDs(ctx context.Context, ids []string) (map[string]models.User, error)
}

type ProductServiceGrpc interface {
	FetchByIDs(ctx context.Context, ids []string) (map[string]models.Product, error)
}
