package grpc

import (
	"context"

	grpcService "github.com/situmorangbastian/skyros/skyrosgrpc"
	cstmErrs "github.com/situmorangbastian/skyros/userservice/internal/errors"
)

func (g *userGrpcHandler) UserLogin(ctx context.Context, request *grpcService.UserLoginRequest) (*grpcService.UserLoginResponse, error) {
	if request.GetEmail() == "" {
		return nil, cstmErrs.ConflictError("email is required")
	}
	if request.GetPassword() == "" {
		return nil, cstmErrs.ConflictError("password is required")
	}

	res, err := g.userUsecase.Login(ctx, request.Email, request.Password)
	if err != nil {
		return nil, err
	}

	accessToken, err := generateToken(res, g.tokenSecretKey)
	if err != nil {
		return nil, err
	}

	return &grpcService.UserLoginResponse{
		AccessToken: accessToken,
	}, nil
}
