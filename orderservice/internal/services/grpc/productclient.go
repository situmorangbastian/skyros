package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/orderservice/internal/domain/models"
	"github.com/situmorangbastian/skyros/orderservice/internal/helpers"
	"github.com/situmorangbastian/skyros/orderservice/internal/services"
	productpb "github.com/situmorangbastian/skyros/proto/product"
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
	c := productpb.NewProductServiceClient(s.grpcClient)

	r, err := c.GetProducts(ctx, &productpb.GetProductsRequest{
		Ids: ids,
	})
	if err != nil {
		return map[string]models.Product{}, err
	}

	if r.GetResult() == nil {
		return map[string]models.Product{}, nil
	}

	result := map[string]models.Product{}
	for _, res := range r.GetResult() {
		product := models.Product{}
		if err = helpers.CopyStructValue(res, &product); err != nil {
			return map[string]models.Product{}, err
		}
		result[product.ID] = product
	}
	return result, nil
}
