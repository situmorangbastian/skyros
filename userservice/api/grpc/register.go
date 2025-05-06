package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userpb "github.com/situmorangbastian/skyros/proto/user"
	"github.com/situmorangbastian/skyros/userservice/internal/models"
)

func (g *userGrpcHandler) RegisterUser(ctx context.Context, request *userpb.RegisterUserRequest) (*userpb.RegisterUserResponse, error) {
	switch request.GetUserType() {
	case "buyer", "seller":
	default:
		return nil, status.Error(codes.NotFound, "Not Found")
	}

	user := models.User{
		Name:     request.GetName(),
		Email:    request.GetEmail(),
		Address:  request.GetAddress(),
		Password: request.GetPassword(),
		Type:     request.UserType,
	}

	err := g.validators.Validate(user)
	if err != nil {
		return nil, err
	}

	res, err := g.userUsecase.Register(ctx, user)
	if err != nil {
		return nil, err
	}

	accessToken, err := generateToken(res, g.tokenSecretKey)
	if err != nil {
		return nil, err
	}

	return &userpb.RegisterUserResponse{
		AccessToken: accessToken,
	}, nil
}
