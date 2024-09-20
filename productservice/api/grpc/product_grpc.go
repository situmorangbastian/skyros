package grpc

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/situmorangbastian/skyros/productservice/internal/usecase"
	grpcService "github.com/situmorangbastian/skyros/skyrosgrpc"
)

type productHandler struct {
	productUsecase usecase.ProductUsecase
}

func NewProductGrpcServer(productUsecase usecase.ProductUsecase) grpcService.ProductServiceServer {
	return &productHandler{
		productUsecase: productUsecase,
	}
}

func (h productHandler) GetProducts(ctx context.Context, filter *grpcService.ProductFilter) (*grpcService.ProductsResponse, error) {
	response := &grpcService.ProductsResponse{
		Status: &grpcService.Status{
			Code:    int32(http.StatusOK),
			Message: "success",
		},
		Products: map[string]*grpcService.Product{},
	}

	if len(filter.GetIds()) == 0 {
		return response, nil
	}

	products, err := h.productUsecase.FetchByIds(ctx, filter.GetIds())
	if err != nil {
		return &grpcService.ProductsResponse{
			Status: &grpcService.Status{
				Code:    int32(http.StatusInternalServerError),
				Message: errors.Cause(err).Error(),
			},
			Products: map[string]*grpcService.Product{},
		}, nil
	}

	productsGrpc := map[string]*grpcService.Product{}
	for _, product := range products {
		productsGrpc[product.ID] = &grpcService.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       int32(product.Price),
			Seller: &grpcService.User{
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
