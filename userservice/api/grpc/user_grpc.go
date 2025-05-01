package grpc

import (
	"context"

	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcService "github.com/situmorangbastian/skyros/skyrosgrpc"
	pkgErrs "github.com/situmorangbastian/skyros/userservice/internal/errors"
	"github.com/situmorangbastian/skyros/userservice/internal/models"
	"github.com/situmorangbastian/skyros/userservice/internal/usecase"
)

type userGrpcHandler struct {
	userUsecase    usecase.UserUsecase
	tokenSecretKey string
}

func NewUserGrpcServer(userUsecase usecase.UserUsecase, tokenSecretKey string) grpcService.UserServiceServer {
	return &userGrpcHandler{
		userUsecase:    userUsecase,
		tokenSecretKey: tokenSecretKey,
	}
}

func (g *userGrpcHandler) GetUsers(ctx context.Context, filter *grpcService.UserFilter) (*grpcService.UsersResponse, error) {
	response := &grpcService.UsersResponse{
		Status: &grpcService.Status{
			Code:    int32(http.StatusOK),
			Message: "success",
		},
		Users: map[string]*grpcService.User{},
	}

	if len(filter.GetUserIds()) == 0 {
		return response, nil
	}

	users, err := g.userUsecase.FetchUsersByIDs(ctx, filter.GetUserIds())
	if err != nil {
		return &grpcService.UsersResponse{
			Status: &grpcService.Status{
				Code:    int32(http.StatusInternalServerError),
				Message: errors.Cause(err).Error(),
			},
			Users: map[string]*grpcService.User{},
		}, nil
	}

	usersGrpc := map[string]*grpcService.User{}
	for _, user := range users {
		usersGrpc[user.ID] = &grpcService.User{
			Id:      user.ID,
			Name:    user.Name,
			Address: user.Address,
			Email:   user.Email,
			Type:    user.Type,
		}
	}

	response.Users = usersGrpc
	return response, nil
}

func (g *userGrpcHandler) UserLogin(ctx context.Context, request *grpcService.UserLoginRequest) (*grpcService.UserLoginResponse, error) {
	if request.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if request.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	res, err := g.userUsecase.Login(ctx, request.Email, request.Password)
	if err != nil {
		return nil, pkgErrs.GenerateGrpcError(err)
	}

	accessToken, err := generateToken(res, g.tokenSecretKey)
	if err != nil {
		return nil, pkgErrs.GenerateGrpcError(err)
	}

	return &grpcService.UserLoginResponse{
		AccessToken: accessToken,
	}, nil
}

func generateToken(user models.User, secretKey string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["name"] = user.Name
	claims["email"] = user.Email
	claims["address"] = user.Address
	claims["type"] = user.Type
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	accessToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
