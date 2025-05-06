package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userpb "github.com/situmorangbastian/skyros/proto/user"
)

func (g *userGrpcHandler) UserLogin(ctx context.Context, request *userpb.UserLoginRequest) (*userpb.UserLoginResponse, error) {
	if request.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if request.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	res, err := g.userUsecase.Login(ctx, request.Email, request.Password)
	if err != nil {
		return nil, err
	}

	accessToken, err := generateToken(res, g.tokenSecretKey)
	if err != nil {
		return nil, err
	}

	return &userpb.UserLoginResponse{
		AccessToken: accessToken,
	}, nil
}
