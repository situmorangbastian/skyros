package grpc

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/orderservice/internal/domain/models"
	"github.com/situmorangbastian/skyros/orderservice/internal/helpers"
	"github.com/situmorangbastian/skyros/orderservice/internal/services"
	userpb "github.com/situmorangbastian/skyros/proto/user"
)

type userService struct {
	grpcClient *grpc.ClientConn
}

func NewUserService(grpcClient *grpc.ClientConn) services.UserServiceGrpc {
	return userService{
		grpcClient: grpcClient,
	}
}

func (s userService) FetchByIDs(ctx context.Context, ids []string) (map[string]models.User, error) {
	c := userpb.NewUserServiceClient(s.grpcClient)

	r, err := c.GetUsers(ctx, &userpb.UserFilter{
		UserIds: ids,
	})
	if err != nil {
		return map[string]models.User{}, err
	}

	status := r.GetStatus()
	if status.Code != int32(http.StatusOK) {
		return map[string]models.User{}, errors.New(status.GetMessage())
	}

	result := map[string]models.User{}
	if len(r.GetUsers()) > 0 {
		for _, userResponse := range r.GetUsers() {
			user := models.User{}
			if err = helpers.CopyStructValue(userResponse, &user); err != nil {
				return map[string]models.User{}, err
			}

			result[user.ID] = user
		}
	}

	return result, nil
}
