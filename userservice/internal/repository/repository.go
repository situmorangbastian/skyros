package repository

import (
	"context"
	"errors"

	"github.com/situmorangbastian/skyros/userservice/internal/models"
)

var (
	ErrNotFound = errors.New("not found")
)

type UserRepository interface {
	Register(ctx context.Context, user models.User) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	FetchUsersByIDs(ctx context.Context, ids []string) (map[string]models.User, error)
}
