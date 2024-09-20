package grpc

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	grpcService "github.com/situmorangbastian/skyros/skyrosgrpc"
	"github.com/situmorangbastian/skyros/userservice/internal/usecase"
)

type userGrpcHandler struct {
	userUsecase usecase.UserUsecase
}

func NewUserGrpcServer(userUsecase usecase.UserUsecase) grpcService.UserServiceServer {
	return &userGrpcHandler{
		userUsecase: userUsecase,
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
