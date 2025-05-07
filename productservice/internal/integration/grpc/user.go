package grpc

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/productservice/internal/helpers"
	"github.com/situmorangbastian/skyros/productservice/internal/integration"
	"github.com/situmorangbastian/skyros/productservice/internal/models"
	userpb "github.com/situmorangbastian/skyros/proto/user"
)

type userClient struct {
	grpcClient *grpc.ClientConn
}

func NewUserIntegrationClient(grpcClient *grpc.ClientConn) integration.UserClient {
	return &userClient{
		grpcClient: grpcClient,
	}
}

func (s *userClient) FetchByIDs(ctx context.Context, ids []string) (map[string]models.User, error) {
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
