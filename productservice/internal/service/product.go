package service

import (
	"context"

	"github.com/situmorangbastian/skyros/productservice/internal/models"
	"github.com/situmorangbastian/skyros/productservice/internal/usecase"
	"github.com/situmorangbastian/skyros/productservice/internal/validation"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	userpb "github.com/situmorangbastian/skyros/proto/user"
)

type handler struct {
	productUsecase usecase.ProductUsecase
	validators     validation.CustomValidator
}

func NewProductService(productUsecase usecase.ProductUsecase, validators validation.CustomValidator) productpb.ProductServiceServer {
	return &handler{
		productUsecase: productUsecase,
		validators:     validators,
	}
}

func (h *handler) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.Product, error) {
	product, err := h.productUsecase.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &productpb.Product{
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
	}, nil
}

func (h *handler) GetProducts(ctx context.Context, filter *productpb.GetProductsRequest) (*productpb.GetProductsResponse, error) {
	if len(filter.GetIds()) != 0 {
		products, err := h.productUsecase.FetchByIds(ctx, filter.GetIds())
		if err != nil {
			return nil, err
		}

		result := []*productpb.Product{}
		for _, product := range products {
			result = append(result, &productpb.Product{
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
			})
		}

		return &productpb.GetProductsResponse{
			Result: result,
		}, nil
	}

	limit := filter.GetLimit()
	if limit == 0 {
		limit = 20
	}

	products, err := h.productUsecase.Fetch(ctx, models.ProductFilter{
		PageSize: int(limit),
		Page:     int(filter.GetOffset()),
		Search:   filter.GetSearch(),
	})
	if err != nil {
		return nil, err
	}

	result := []*productpb.Product{}
	for _, product := range products {
		result = append(result, &productpb.Product{
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
		})
	}

	return &productpb.GetProductsResponse{
		Result: result,
	}, nil
}

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
