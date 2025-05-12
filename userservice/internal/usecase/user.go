package usecase

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if err == repository.ErrNotFound {
			return models.User{}, status.Error(codes.NotFound, "user not found")
		}
		return models.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return models.User{}, status.Error(codes.NotFound, "user not found")
	}

	return user, nil
}

func (u *userUsecase) Register(ctx context.Context, user models.User) (models.User, error) {
	currentUser, err := u.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil && err != repository.ErrNotFound {
		return models.User{}, err
	}

	if currentUser.Email == user.Email {
		return models.User{}, status.Error(codes.AlreadyExists, "email already exist")
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, errors.Wrap(err, "user.service.register: hash password")
	}

	user.Password = string(hashPassword)

	result, err := u.userRepo.Register(ctx, user)
	if err != nil {
		return models.User{}, err
	}

	return result, nil
}

func (u *userUsecase) FetchUsersByIDs(ctx context.Context, ids []string) (map[string]models.User, error) {
	users, err := u.userRepo.FetchUsersByIDs(ctx, ids)
	if err != nil {
		return map[string]models.User{}, err
	}

	return users, nil
}
