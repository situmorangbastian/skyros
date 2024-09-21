package usecase

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	internalErr "github.com/situmorangbastian/skyros/userservice/internal/errors"
	"github.com/situmorangbastian/skyros/userservice/internal/models"
	"github.com/situmorangbastian/skyros/userservice/internal/repository"
)

type UserUsecase interface {
	Login(ctx context.Context, email, password string) (models.User, error)
	Register(ctx context.Context, user models.User) (models.User, error)
	FetchUsersByIDs(ctx context.Context, ids []string) (map[string]models.User, error)
}

type userUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(userRepo repository.UserRepository) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
	}
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (models.User, error) {
	user, err := u.userRepo.GetUser(ctx, email)
	if err != nil {
		return models.User{}, errors.Wrap(err, "user.service.login: get user by email repo")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return models.User{}, internalErr.NotFoundError("user not found")
	}

	return user, nil
}

func (u *userUsecase) Register(ctx context.Context, user models.User) (models.User, error) {
	currentUser, err := u.userRepo.GetUser(ctx, user.Email)
	if err != nil {
		switch errors.Cause(err).(type) {
		case internalErr.NotFoundError:
		default:
			return models.User{}, errors.Wrap(err, "user.service.register: get user by email")
		}
	}

	if currentUser.Email == user.Email {
		return models.User{}, internalErr.ConflictError("email already exist")
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, errors.Wrap(err, "user.service.register: hash password")
	}

	user.Password = string(hashPassword)

	result, err := u.userRepo.Register(ctx, user)
	if err != nil {
		return models.User{}, errors.Wrap(err, "user.service.register: register repo")
	}

	return result, nil
}

func (u *userUsecase) FetchUsersByIDs(ctx context.Context, ids []string) (map[string]models.User, error) {
	users, err := u.userRepo.FetchUsersByIDs(ctx, ids)
	if err != nil {
		return map[string]models.User{}, errors.Wrap(err, "user.service.fetch: get users by ids")
	}

	return users, nil
}
