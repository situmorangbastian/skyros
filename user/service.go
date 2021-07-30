package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/situmorangbastian/skyros"
)

type service struct {
	repo skyros.UserRepository
}

func NewService(repo skyros.UserRepository) skyros.UserService {
	return service{
		repo: repo,
	}
}

func (s service) Login(ctx context.Context, email, password string) (skyros.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return skyros.User{}, errors.Wrap(err, "user.service.login: get user by email repo")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return skyros.User{}, skyros.ErrorNotFound("user not found")
	}

	return user, nil
}

func (s service) Register(ctx context.Context, user skyros.User) (skyros.User, error) {
	currentUser, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		switch errors.Cause(err).(type) {
		case skyros.ErrorNotFound:
		default:
			return skyros.User{}, errors.Wrap(err, "user.service.register: get user by email")
		}
	}

	if currentUser.Email == user.Email {
		return skyros.User{}, skyros.ConflictError("email already exist")
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return skyros.User{}, errors.Wrap(err, "user.service.register: hash password")
	}

	user.ID = uuid.New().String()
	user.Password = string(hashPassword)

	result, err := s.repo.Register(ctx, user)
	if err != nil {
		return skyros.User{}, errors.Wrap(err, "user.service.register: register repo")
	}

	return result, nil
}
