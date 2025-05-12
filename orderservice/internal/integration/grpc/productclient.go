package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/orderservice/internal/integration"
	"github.com/situmorangbastian/skyros/orderservice/internal/models"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	svcutils "github.com/situmorangbastian/skyros/serviceutils"
)

type productClient struct {
	grpcClient *grpc.ClientConn
}

func NewProductClient(grpcClient *grpc.ClientConn) integration.ProductClient {
	return &productClient{
		grpcClient: grpcClient,
	}
}

func (pc *productClient) FetchByIDs(ctx context.Context, ids []string) (map[string]models.Product, error) {
	c := productpb.NewProductServiceClient(pc.grpcClient)

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
		if err = svcutils.CopyStructValue(res, &product); err != nil {
			return map[string]models.Product{}, err
		}
		result[product.ID] = product
	}
	return result, nil
}
