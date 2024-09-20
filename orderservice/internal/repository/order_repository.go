package repository

import (
	"context"

	"github.com/situmorangbastian/skyros/orderservice/internal/domain/models"
)

type OrderRepository interface {
	Store(ctx context.Context, order models.Order) (models.Order, error)
	Fetch(ctx context.Context, filter models.Filter) ([]models.Order, string, error)
	PatchStatus(ctx context.Context, ID string, status int) error
}
