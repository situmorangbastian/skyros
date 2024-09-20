package grpc

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/orderservice/internal/domain/models"
	"github.com/situmorangbastian/skyros/orderservice/internal/helpers"
	"github.com/situmorangbastian/skyros/orderservice/internal/services"
	grpcService "github.com/situmorangbastian/skyros/skyrosgrpc"
)

type productService struct {
	grpcClient *grpc.ClientConn
}

func NewProductService(grpcClient *grpc.ClientConn) services.ProductServiceGrpc {
	return productService{
		grpcClient: grpcClient,
	}
}

func (s productService) FetchByIDs(ctx context.Context, ids []string) (map[string]models.Product, error) {
	c := grpcService.NewProductServiceClient(s.grpcClient)

	r, err := c.GetProducts(ctx, &grpcService.ProductFilter{
		Ids: ids,
	})
	if err != nil {
		return map[string]models.Product{}, err
	}

	status := r.GetStatus()
	if status.Code != int32(http.StatusOK) {
		return map[string]models.Product{}, errors.New(status.GetMessage())
	}

	result := map[string]models.Product{}
	if len(r.GetProducts()) > 0 {
		for _, productResponse := range r.GetProducts() {
			product := models.Product{}
			if err = helpers.CopyStructValue(productResponse, &product); err != nil {
				return map[string]models.Product{}, err
			}

			result[product.ID] = product
		}
	}

	return result, nil
}
