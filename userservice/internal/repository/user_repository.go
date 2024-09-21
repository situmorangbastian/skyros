package repository

import (
	"context"

	"github.com/situmorangbastian/skyros/userservice/internal/models"
)

type UserRepository interface {
	Register(ctx context.Context, user models.User) (models.User, error)
	GetUser(ctx context.Context, identifier string) (models.User, error)
	FetchUsersByIDs(ctx context.Context, ids []string) (map[string]models.User, error)
}
