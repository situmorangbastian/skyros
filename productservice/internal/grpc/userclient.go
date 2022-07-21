package grpc

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/productservice"
	grpcService "github.com/situmorangbastian/skyros/skyrosgrpc"
)

type userService struct {
	grpcClientConn *grpc.ClientConn
}

func NewUserService(grpcClientConn *grpc.ClientConn) productservice.UserServiceGrpc {
	return userService{
		grpcClientConn: grpcClientConn,
	}
}

func (s userService) FetchByIDs(ctx context.Context, ids []string) (map[string]productservice.User, error) {
	c := grpcService.NewUserServiceClient(s.grpcClientConn)

	r, err := c.GetUsers(ctx, &grpcService.UserFilter{
		UserIds: ids,
	})
	if err != nil {
		return map[string]productservice.User{}, err
	}

	status := r.GetStatus()
	if status.Code != int32(http.StatusOK) {
		return map[string]productservice.User{}, errors.New(status.GetMessage())
	}

	result := map[string]productservice.User{}
	if len(r.GetUsers()) > 0 {
		for _, userResponse := range r.GetUsers() {
			user := productservice.User{}
			if err = productservice.CopyStructValue(userResponse, &user); err != nil {
				return map[string]productservice.User{}, err
			}

			result[user.ID] = user
		}
	}

	return result, nil
}
