package grpc

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/orderservice/internal/integration"
	"github.com/situmorangbastian/skyros/orderservice/internal/models"
	userpb "github.com/situmorangbastian/skyros/proto/user"
	svcutils "github.com/situmorangbastian/skyros/serviceutils"
)

type userClient struct {
	grpcClient *grpc.ClientConn
}

func NewUserClient(grpcClient *grpc.ClientConn) integration.UserClient {
	return &userClient{
		grpcClient: grpcClient,
	}
}

func (uc *userClient) FetchByIDs(ctx context.Context, ids []string) (map[string]models.User, error) {
	c := userpb.NewUserServiceClient(uc.grpcClient)

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
			if err = svcutils.CopyStructValue(userResponse, &user); err != nil {
				return map[string]models.User{}, err
			}

			result[user.ID] = user
		}
	}

	return result, nil
}
