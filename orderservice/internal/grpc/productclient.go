package grpc

import (
	"context"
	"errors"
	"net/http"

	grpcService "github.com/situmorangbastian/skyrosgrpc"
	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/orderservice"
)

type productService struct {
	grpcClientConn *grpc.ClientConn
}

func NewProductService(grpcClientConn *grpc.ClientConn) orderservice.ProductServiceGrpc {
	return productService{
		grpcClientConn: grpcClientConn,
	}
}

func (s productService) FetchByIDs(ctx context.Context, ids []string) (map[string]orderservice.Product, error) {
	c := grpcService.NewProductServiceClient(s.grpcClientConn)

	r, err := c.GetProducts(ctx, &grpcService.ProductFilter{
		Ids: ids,
	})
	if err != nil {
		return map[string]orderservice.Product{}, err
	}

	status := r.GetStatus()
	if status.Code != int32(http.StatusOK) {
		return map[string]orderservice.Product{}, errors.New(status.GetMessage())
	}

	result := map[string]orderservice.Product{}
	if len(r.GetProducts()) > 0 {
		for _, productResponse := range r.GetProducts() {
			product := orderservice.Product{}
			if err = orderservice.CopyStructValue(productResponse, &product); err != nil {
				return map[string]orderservice.Product{}, err
			}

			result[product.ID] = product
		}
	}

	return result, nil
}
