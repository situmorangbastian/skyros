package services

import (
	"context"

	"github.com/situmorangbastian/skyros/productservice/internal/models"
)

type UserGrpcService interface {
	FetchByIDs(ctx context.Context, ids []string) (map[string]models.User, error)
}
