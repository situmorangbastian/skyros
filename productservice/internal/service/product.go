package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/situmorangbastian/skyros/productservice/internal/models"
	"github.com/situmorangbastian/skyros/productservice/internal/usecase"
	productpb "github.com/situmorangbastian/skyros/proto/product"
	userpb "github.com/situmorangbastian/skyros/proto/user"
	"github.com/situmorangbastian/skyros/serviceutils"
)

type handler struct {
	productUsecase usecase.ProductUsecase
	validators     serviceutils.CustomValidator
}

func NewProductService(productUsecase usecase.ProductUsecase, validators serviceutils.CustomValidator) productpb.ProductServiceServer {
	return &handler{
		productUsecase: productUsecase,
		validators:     validators,
	}
}

func (h *handler) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.Product, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "internal.service.product.GetProduct").Logger()
	log.Info().Msg("request received")

	// ids are UUID columns; a malformed id would otherwise surface as a 500
	if uuid.Validate(req.GetId()) != nil {
		return nil, status.Error(codes.NotFound, "product not found")
	}

	product, err := h.productUsecase.Get(ctx, req.GetId())
	if err != nil {
		log.Error().Err(err).Msg("failed get product")
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
			Type:    string(product.Seller.Type),
		},
	}, nil
}

func (h *handler) GetProducts(ctx context.Context, filter *productpb.GetProductsRequest) (*productpb.GetProductsResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "internal.service.product.GetProducts").Logger()
	log.Info().Msg("request received")

	if len(filter.GetIds()) != 0 {
		for _, id := range filter.GetIds() {
			if uuid.Validate(id) != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid product id")
			}
		}

		products, err := h.productUsecase.FetchByIds(ctx, filter.GetIds())
		if err != nil {
			log.Error().Err(err).Msg("failed get products")
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
					Type:    string(product.Seller.Type),
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
		log.Error().Err(err).Msg("failed get products")
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
				Type:    string(product.Seller.Type),
			},
		})
	}

	return &productpb.GetProductsResponse{
		Result: result,
	}, nil
}

func (h *handler) StoreProduct(ctx context.Context, request *productpb.StoreProductRequest) (*productpb.Product, error) {
	log := zerolog.Ctx(ctx).With().Str("func", "internal.service.product.StoreProduct").Logger()
	log.Info().Msg("request received")

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
		log.Error().Err(err).Msg("failed store products")
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
			Type:    string(product.Seller.Type),
		},
	}

	return productsGrpc, nil
}
