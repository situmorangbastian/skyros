package grpc

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
	grpc "github.com/situmorangbastian/skyros/skyrosgrpc"

	"github.com/situmorangbastian/skyros/productservice"
)

type productHandler struct {
	service productservice.ProductService
}

func NewProductGrpcServer(service productservice.ProductService) grpc.ProductServiceServer {
	return &productHandler{
		service: service,
	}
}

func (h productHandler) GetProducts(ctx context.Context, filter *grpc.ProductFilter) (*grpc.ProductsResponse, error) {
	response := &grpc.ProductsResponse{
		Status: &grpc.Status{
			Code:    int32(http.StatusOK),
			Message: "success",
		},
		Products: map[string]*grpc.Product{},
	}

	if len(filter.GetIds()) == 0 {
		return response, nil
	}

	products, err := h.service.FetchByIds(ctx, filter.GetIds())
	if err != nil {
		return &grpc.ProductsResponse{
			Status: &grpc.Status{
				Code:    int32(http.StatusInternalServerError),
				Message: errors.Cause(err).Error(),
			},
			Products: map[string]*grpc.Product{},
		}, nil
	}

	productsGrpc := map[string]*grpc.Product{}
	for _, product := range products {
		productsGrpc[product.ID] = &grpc.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       int32(product.Price),
			Seller: &grpc.User{
				Id:      product.Seller.ID,
				Email:   product.Seller.Email,
				Name:    product.Seller.Name,
				Address: product.Seller.Address,
				Type:    product.Seller.Type,
			},
		}
	}

	response.Products = productsGrpc
	return response, nil
}
