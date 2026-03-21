package grpc

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	userpb "github.com/situmorangbastian/skyros/proto/user"
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

func (uc *userClient) FetchByIDs(ctx context.Context, ids []string) (map[string]auth.Claims, error) {
	c := userpb.NewUserServiceClient(uc.grpcClient)

	r, err := c.GetUsers(ctx, &userpb.UserFilter{
		UserIds: ids,
	})
	if err != nil {
		return map[string]auth.Claims{}, err
	}

	status := r.GetStatus()
	if status.Code != int32(http.StatusOK) {
		return map[string]auth.Claims{}, errors.New(status.GetMessage())
	}

	result := map[string]auth.Claims{}
	if len(r.GetUsers()) > 0 {
		for _, userResponse := range r.GetUsers() {
			user := auth.ToAuthClaims(userResponse)
			result[user.ID] = user
		}
	}

	return result, nil
}
