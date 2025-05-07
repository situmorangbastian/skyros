package grpc

import (
	"github.com/situmorangbastian/skyros/productservice/api/validators"
	"github.com/situmorangbastian/skyros/productservice/internal/usecase"
	productpb "github.com/situmorangbastian/skyros/proto/product"
)

type handler struct {
	productUsecase usecase.ProductUsecase
	validators     validators.CustomValidator
}

func NewProductGrpcServer(productUsecase usecase.ProductUsecase, validators validators.CustomValidator) productpb.ProductServiceServer {
	return &handler{
		productUsecase: productUsecase,
		validators:     validators,
	}
}
