package service

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commonpb "github.com/situmorangbastian/skyros/proto/common"
	userpb "github.com/situmorangbastian/skyros/proto/user"
	"github.com/situmorangbastian/skyros/userservice/internal/models"
	"github.com/situmorangbastian/skyros/userservice/internal/usecase"
	"github.com/situmorangbastian/skyros/userservice/internal/validation"
)

type service struct {
	userUsecase    usecase.UserUsecase
	tokenSecretKey string
	validators     validation.CustomValidator
	logger         zerolog.Logger
}

func NewUserService(userUsecase usecase.UserUsecase, tokenSecretKey string, validators validation.CustomValidator, logger zerolog.Logger) userpb.UserServiceServer {
	return &service{
		userUsecase:    userUsecase,
		tokenSecretKey: tokenSecretKey,
		validators:     validators,
		logger:         logger,
	}
}

func (s *service) GetUsers(ctx context.Context, filter *userpb.UserFilter) (*userpb.UsersResponse, error) {
	log := s.logger.With().Str("func", "internal.service.user.GetUsers").Logger()

	response := &userpb.UsersResponse{
		Status: &commonpb.Status{
			Code:    int32(http.StatusOK),
			Message: "success",
		},
		Users: map[string]*userpb.User{},
	}

	if len(filter.GetUserIds()) == 0 {
		return response, nil
	}

	users, err := s.userUsecase.FetchUsersByIDs(ctx, filter.GetUserIds())
	if err != nil {
		log.Error().Err(err).Msg("failed FetchUsersByIDs")
		return &userpb.UsersResponse{
			Status: &commonpb.Status{
				Code:    int32(http.StatusInternalServerError),
				Message: errors.Cause(err).Error(),
			},
			Users: map[string]*userpb.User{},
		}, nil
	}

	usersGrpc := map[string]*userpb.User{}
	for _, user := range users {
		usersGrpc[user.ID] = &userpb.User{
			Id:      user.ID,
			Name:    user.Name,
			Address: user.Data.Address,
			Email:   user.Email,
			Type:    user.Data.Type,
		}
	}

	response.Users = usersGrpc
	return response, nil
}

func (s *service) UserLogin(ctx context.Context, request *userpb.UserLoginRequest) (*userpb.UserLoginResponse, error) {
	log := s.logger.With().Str("func", "internal.service.user.UserLogin").Logger()

	if request.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if request.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	loginRequest := models.UserLoginRequest{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
	}

	err := s.validators.Validate(loginRequest)
	if err != nil {
		return nil, err
	}

	res, err := s.userUsecase.Login(ctx, request.Email, request.Password)
	if err != nil {
		log.Error().Err(err).Msg("failed Login")
		return nil, err
	}

	accessToken, err := generateToken(res, s.tokenSecretKey, log)
	if err != nil {
		log.Error().Err(err).Msg("failed generateToken")
		return nil, err
	}

	return &userpb.UserLoginResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *service) RegisterUser(ctx context.Context, request *userpb.RegisterUserRequest) (*userpb.RegisterUserResponse, error) {
	log := s.logger.With().Str("func", "internal.service.user.RegisterUser").Logger()

	switch request.GetUserType() {
	case "buyer", "seller":
	default:
		return nil, status.Error(codes.NotFound, "Not Found")
	}

	user := models.User{
		Name:     request.GetName(),
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
		Data: models.UserData{
			Address: request.GetAddress(),
			Type:    request.UserType,
		},
	}

	err := s.validators.Validate(user)
	if err != nil {
		return nil, err
	}

	res, err := s.userUsecase.Register(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("failed Register")
		return nil, err
	}

	accessToken, err := generateToken(res, s.tokenSecretKey, log)
	if err != nil {
		log.Error().Err(err).Msg("failed generateToken")
		return nil, err
	}

	return &userpb.RegisterUserResponse{
		AccessToken: accessToken,
	}, nil
}

func generateToken(user models.User, secretKey string, log zerolog.Logger) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["name"] = user.Name
	claims["email"] = user.Email
	claims["address"] = user.Data.Address
	claims["type"] = user.Data.Type
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	accessToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Error().Err(err).Msg("failed token.SignedString")
		return "", status.Error(codes.Internal, "Internal Server Error")
	}

	return accessToken, nil
}
