package usecase

import (
	"context"

	"github.com/rs/zerolog"
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
	logger   zerolog.Logger
}

func NewUserUsecase(userRepo repository.UserRepository, logger zerolog.Logger) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (models.User, error) {
	log := u.logger.With().Str("func", "internal.usecase.user.Login").Logger()

	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if err == repository.ErrNotFound {
			return models.User{}, status.Error(codes.NotFound, "user not found")
		}
		log.Error().Err(err).Msg("failed GetUserByEmail")
		return models.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return models.User{}, status.Error(codes.NotFound, "user not found")
	}

	return user, nil
}

func (u *userUsecase) Register(ctx context.Context, user models.User) (models.User, error) {
	log := u.logger.With().Str("func", "internal.usecase.user.Register").Logger()

	currentUser, err := u.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil && err != repository.ErrNotFound {
		log.Error().Err(err).Msg("failed GetUserByEmail")
		return models.User{}, err
	}

	if currentUser.Email == user.Email {
		return models.User{}, status.Error(codes.AlreadyExists, "email already exist")
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("failed GenerateFromPassword")
		return models.User{}, status.Error(codes.Internal, "Internal Server Error")
	}

	user.Password = string(hashPassword)

	result, err := u.userRepo.Register(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("failed Register")
		return models.User{}, err
	}

	return result, nil
}

func (u *userUsecase) FetchUsersByIDs(ctx context.Context, ids []string) (map[string]models.User, error) {
	log := u.logger.With().Str("func", "internal.usecase.user.FetchUsersByIDs").Logger()

	users, err := u.userRepo.FetchUsersByIDs(ctx, ids)
	if err != nil {
		log.Error().Err(err).Msg("failed FetchUsersByIDs")
		return map[string]models.User{}, err
	}

	return users, nil
}
