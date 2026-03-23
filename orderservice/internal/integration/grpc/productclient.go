package grpc

import (
	"context"

	"github.com/situmorangbastian/skyros/orderservice/internal/integration"
	"github.com/situmorangbastian/skyros/orderservice/internal/models"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	userpb "github.com/situmorangbastian/skyros/proto/user"
	"github.com/situmorangbastian/skyros/serviceutils/auth"
)

type productClient struct {
	productSvcClient productpb.ProductServiceClient
}

func NewProductClient(productSvcClient productpb.ProductServiceClient) integration.ProductClient {
	return &productClient{
		productSvcClient: productSvcClient,
	}
}

func (pc *productClient) FetchByIDs(ctx context.Context, ids []string) (map[string]models.Product, error) {
	resp, err := pc.productSvcClient.GetProducts(ctx, &productpb.GetProductsRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}

	return toProductMap(resp.GetResult()), nil
}

func toProductMap(products []*productpb.Product) map[string]models.Product {
	result := make(map[string]models.Product, len(products))
	for _, p := range products {
		product := toProductModel(p)
		result[product.ID] = product
	}
	return result
}

func toProductModel(p *productpb.Product) models.Product {
	return models.Product{
		ID:          p.GetId(),
		Name:        p.GetName(),
		Description: p.GetDescription(),
		Price:       p.GetPrice(),
		Seller:      toSellerClaims(p.GetSeller()),
	}
}

func toSellerClaims(s *userpb.User) auth.Claims {
	if s == nil {
		return auth.Claims{}
	}
	return auth.Claims{
		ID:      s.GetId(),
		Email:   s.GetEmail(),
		Name:    s.GetName(),
		Address: s.GetAddress(),
	}
}
