package user

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/situmorangbastian/skyros/userservice"
)

type service struct {
	repo userservice.UserRepository
}

func NewService(repo userservice.UserRepository) userservice.UserService {
	return service{
		repo: repo,
	}
}

func (s service) Login(ctx context.Context, email, password string) (userservice.User, error) {
	user, err := s.repo.GetUser(ctx, email)
	if err != nil {
		return userservice.User{}, errors.Wrap(err, "user.service.login: get user by email repo")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return userservice.User{}, userservice.ErrorNotFound("user not found")
	}

	return user, nil
}

func (s service) Register(ctx context.Context, user userservice.User) (userservice.User, error) {
	currentUser, err := s.repo.GetUser(ctx, user.Email)
	if err != nil {
		switch errors.Cause(err).(type) {
		case userservice.ErrorNotFound:
		default:
			return userservice.User{}, errors.Wrap(err, "user.service.register: get user by email")
		}
	}

	if currentUser.Email == user.Email {
		return userservice.User{}, userservice.ConflictError("email already exist")
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return userservice.User{}, errors.Wrap(err, "user.service.register: hash password")
	}

	user.Password = string(hashPassword)

	result, err := s.repo.Register(ctx, user)
	if err != nil {
		return userservice.User{}, errors.Wrap(err, "user.service.register: register repo")
	}

	return result, nil
}
