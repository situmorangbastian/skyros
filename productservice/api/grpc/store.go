package grpc

import (
	"context"

	"github.com/situmorangbastian/skyros/productservice/internal/models"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	userpb "github.com/situmorangbastian/skyros/proto/user"
)

func (h *handler) StoreProduct(ctx context.Context, request *productpb.StoreProductRequest) (*productpb.Product, error) {
	productReq := models.Product{
		Name:        request.GetName(),
		Description: request.GetDescription(),
		Price:       int64(request.Price),
	}

	err := h.validators.Validate(productReq)
	if err != nil {
		return nil, err
	}

	product, err := h.productUsecase.Store(ctx, productReq)
	if err != nil {
		return nil, err
	}

	productsGrpc := &productpb.Product{
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

	return productsGrpc, nil
}
