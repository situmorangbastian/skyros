package integration

import (
	"context"

	"github.com/situmorangbastian/skyros/orderservice/internal/models"
)

type UserClient interface {
	FetchByIDs(ctx context.Context, ids []string) (map[string]models.User, error)
}
