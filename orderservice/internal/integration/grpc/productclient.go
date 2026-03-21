package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/situmorangbastian/skyros/orderservice/internal/integration"
	"github.com/situmorangbastian/skyros/orderservice/internal/models"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	"github.com/situmorangbastian/skyros/serviceutils/auth"
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
		product := toProductModel(res)
		result[product.ID] = product
	}
	return result, nil
}

func toProductModel(p *productpb.Product) models.Product {
	return models.Product{
		ID:          p.GetId(),
		Name:        p.GetName(),
		Description: p.GetDescription(),
		Price:       p.GetPrice(),
		Seller: auth.Claims{
			ID:      p.GetSeller().Id,
			Email:   p.GetSeller().Email,
			Name:    p.GetSeller().Name,
			Address: p.GetSeller().Address,
		},
	}
}
