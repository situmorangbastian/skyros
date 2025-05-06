package grpc

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/situmorangbastian/skyros/productservice/internal/usecase"
	commonpb "github.com/situmorangbastian/skyros/proto/common"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	userpb "github.com/situmorangbastian/skyros/proto/user"
)

type productHandler struct {
	productUsecase usecase.ProductUsecase
}

func NewProductGrpcServer(productUsecase usecase.ProductUsecase) productpb.ProductServiceServer {
	return &productHandler{
		productUsecase: productUsecase,
	}
}

func (h productHandler) GetProducts(ctx context.Context, filter *productpb.ProductFilter) (*productpb.ProductsResponse, error) {
	response := &productpb.ProductsResponse{
		Status: &commonpb.Status{
			Code:    int32(http.StatusOK),
			Message: "success",
		},
		Products: map[string]*productpb.Product{},
	}

	if len(filter.GetIds()) == 0 {
		return response, nil
	}

	products, err := h.productUsecase.FetchByIds(ctx, filter.GetIds())
	if err != nil {
		return &productpb.ProductsResponse{
			Status: &commonpb.Status{
				Code:    int32(http.StatusInternalServerError),
				Message: errors.Cause(err).Error(),
			},
			Products: map[string]*productpb.Product{},
		}, nil
	}

	productsGrpc := map[string]*productpb.Product{}
	for _, product := range products {
		productsGrpc[product.ID] = &productpb.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       int32(product.Price),
			Seller: &userpb.User{
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
