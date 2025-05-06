package grpc

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	commonpb "github.com/situmorangbastian/skyros/proto/common"
	userpb "github.com/situmorangbastian/skyros/proto/user"
)

func (g *userGrpcHandler) GetUsers(ctx context.Context, filter *userpb.UserFilter) (*userpb.UsersResponse, error) {
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

	users, err := g.userUsecase.FetchUsersByIDs(ctx, filter.GetUserIds())
	if err != nil {
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
			Address: user.Address,
			Email:   user.Email,
			Type:    user.Type,
		}
	}

	response.Users = usersGrpc
	return response, nil
}
