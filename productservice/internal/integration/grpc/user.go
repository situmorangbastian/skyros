package grpc

import (
	"context"
	"errors"
	"net/http"

	"github.com/situmorangbastian/skyros/proto/common"
	userpb "github.com/situmorangbastian/skyros/proto/user"
	"github.com/situmorangbastian/skyros/serviceutils/auth"
)

type userClient struct {
	userSvcClient userpb.UserServiceClient
}

func NewUserClient(userSvcClient userpb.UserServiceClient) auth.UserClient {
	return &userClient{
		userSvcClient: userSvcClient,
	}
}

func (uc *userClient) FetchByIDs(ctx context.Context, ids []string) (map[string]auth.Claims, error) {
	resp, err := uc.userSvcClient.GetUsers(ctx, &userpb.UserFilter{
		UserIds: ids,
	})
	if err != nil {
		return nil, err
	}

	if err := validateStatus(resp.GetStatus()); err != nil {
		return nil, err
	}

	return toClaimsMap(resp.GetUsers()), nil
}

func validateStatus(status *common.Status) error {
	if status == nil || status.Code == int32(http.StatusOK) {
		return nil
	}
	return errors.New(status.GetMessage())
}

func toClaimsMap(users map[string]*userpb.User) map[string]auth.Claims {
	result := make(map[string]auth.Claims, len(users))
	for _, u := range users {
		claim := auth.ToAuthClaims(u)
		result[claim.ID] = claim
	}
	return result
}
