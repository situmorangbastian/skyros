package grpc

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	userpb "github.com/situmorangbastian/skyros/proto/user"
	svcutils "github.com/situmorangbastian/skyros/serviceutils"
	"github.com/situmorangbastian/skyros/serviceutils/auth"
)

type userClient struct {
	grpcClient *grpc.ClientConn
}

func NewUserClient(grpcClient *grpc.ClientConn) auth.UserClient {
	return &userClient{
		grpcClient: grpcClient,
	}
}

func (uc *userClient) FetchByIDs(ctx context.Context, ids []string) (map[string]auth.User, error) {
	c := userpb.NewUserServiceClient(uc.grpcClient)

	r, err := c.GetUsers(ctx, &userpb.UserFilter{
		UserIds: ids,
	})
	if err != nil {
		return map[string]auth.User{}, err
	}

	status := r.GetStatus()
	if status.Code != int32(http.StatusOK) {
		return map[string]auth.User{}, errors.New(status.GetMessage())
	}

	result := map[string]auth.User{}
	if len(r.GetUsers()) > 0 {
		for _, userResponse := range r.GetUsers() {
			user := auth.User{}
			if err = svcutils.CopyStructValue(userResponse, &user); err != nil {
				return map[string]auth.User{}, err
			}

			result[user.ID] = user
		}
	}

	return result, nil
}
