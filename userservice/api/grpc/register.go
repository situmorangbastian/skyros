package grpc

import (
	"context"

	grpcService "github.com/situmorangbastian/skyros/skyrosgrpc"
	cstmErrs "github.com/situmorangbastian/skyros/userservice/internal/errors"
	"github.com/situmorangbastian/skyros/userservice/internal/models"
)

func (g *userGrpcHandler) RegisterUser(ctx context.Context, request *grpcService.RegisterUserRequest) (*grpcService.RegisterUserResponse, error) {
	switch request.GetUserType() {
	case "buyer", "seller":
	default:
		return nil, cstmErrs.NotFoundError("Not Found")
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

	return &grpcService.RegisterUserResponse{
		AccessToken: accessToken,
	}, nil
}
