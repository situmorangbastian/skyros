package repository

import (
	"context"
	"errors"

	"github.com/situmorangbastian/skyros/orderservice/internal/models"
)

var (
	ErrNotFound = errors.New("not found")
)

type OrderRepository interface {
	Store(ctx context.Context, order models.Order) (models.Order, error)
	Fetch(ctx context.Context, filter models.Filter) ([]models.Order, error)
	PatchStatus(ctx context.Context, ID string, status int) error
}
