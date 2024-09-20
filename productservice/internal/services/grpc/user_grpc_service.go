package grpc

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/productservice/internal/helpers"
	"github.com/situmorangbastian/skyros/productservice/internal/models"
	"github.com/situmorangbastian/skyros/productservice/internal/services"
	grpcService "github.com/situmorangbastian/skyros/skyrosgrpc"
)

type userGrpcSvcImpl struct {
	grpcClient *grpc.ClientConn
}

func NewUserService(grpcClient *grpc.ClientConn) services.UserGrpcService {
	return &userGrpcSvcImpl{
		grpcClient: grpcClient,
	}
}

func (s *userGrpcSvcImpl) FetchByIDs(ctx context.Context, ids []string) (map[string]models.User, error) {
	c := grpcService.NewUserServiceClient(s.grpcClient)

	r, err := c.GetUsers(ctx, &grpcService.UserFilter{
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
