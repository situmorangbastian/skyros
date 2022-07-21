package grpc

import (
	"context"
	"errors"
	"net/http"

	grpcService "github.com/situmorangbastian/skyrosgrpc"
	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/orderservice"
)

type userService struct {
	grpcClientConn *grpc.ClientConn
}

func NewUserService(grpcClientConn *grpc.ClientConn) orderservice.UserServiceGrpc {
	return userService{
		grpcClientConn: grpcClientConn,
	}
}

func (s userService) FetchByIDs(ctx context.Context, ids []string) (map[string]orderservice.User, error) {
	c := grpcService.NewUserServiceClient(s.grpcClientConn)

	r, err := c.GetUsers(ctx, &grpcService.UserFilter{
		UserIds: ids,
	})
	if err != nil {
		return map[string]orderservice.User{}, err
	}

	status := r.GetStatus()
	if status.Code != int32(http.StatusOK) {
		return map[string]orderservice.User{}, errors.New(status.GetMessage())
	}

	result := map[string]orderservice.User{}
	if len(r.GetUsers()) > 0 {
		for _, userResponse := range r.GetUsers() {
			user := orderservice.User{}
			if err = orderservice.CopyStructValue(userResponse, &user); err != nil {
				return map[string]orderservice.User{}, err
			}

			result[user.ID] = user
		}
	}

	return result, nil
}
