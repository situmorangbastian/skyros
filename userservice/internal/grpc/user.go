package grpc

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/situmorangbastian/skyros/userservice"
	grpc "github.com/situmorangbastian/skyrosgrpc"
)

type userHandler struct {
	service userservice.UserService
}

func NewUserGrpcServer(service userservice.UserService) grpc.UserServiceServer {
	return &userHandler{
		service: service,
	}
}

func (h userHandler) GetUsers(ctx context.Context, filter *grpc.UserFilter) (*grpc.UsersResponse, error) {
	response := &grpc.UsersResponse{
		Status: &grpc.Status{
			Code:    int32(http.StatusOK),
			Message: "success",
		},
		Users: map[string]*grpc.User{},
	}

	if len(filter.GetUserIds()) == 0 {
		return response, nil
	}

	users, err := h.service.FetchUsersByIDs(ctx, filter.GetUserIds())
	if err != nil {
		return &grpc.UsersResponse{
			Status: &grpc.Status{
				Code:    int32(http.StatusInternalServerError),
				Message: errors.Cause(err).Error(),
			},
			Users: map[string]*grpc.User{},
		}, nil
	}

	usersGrpc := map[string]*grpc.User{}
	for _, user := range users {
		usersGrpc[user.ID] = &grpc.User{
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
