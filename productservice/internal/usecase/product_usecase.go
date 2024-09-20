package usecase

import (
	"context"

	"github.com/pkg/errors"

	restCtx "github.com/situmorangbastian/skyros/productservice/api/rest/context"
	internalErr "github.com/situmorangbastian/skyros/productservice/internal/error"
	"github.com/situmorangbastian/skyros/productservice/internal/models"
	"github.com/situmorangbastian/skyros/productservice/internal/repository"
	"github.com/situmorangbastian/skyros/productservice/internal/services"
)

type ProductUsecase interface {
	Store(ctx context.Context, product models.Product) (models.Product, error)
	Get(ctx context.Context, ID string) (models.Product, error)
	Fetch(ctx context.Context, filter models.ProductFilter) ([]models.Product, error)
	FetchByIds(ctx context.Context, ids []string) (map[string]models.Product, error)
}

type usecase struct {
	productRepo repository.ProductRepository
	userGrpcSvc services.UserGrpcService
}

func NewProductUsecase(productRepo repository.ProductRepository, userGrpcSvc services.UserGrpcService) ProductUsecase {
	return &usecase{
		productRepo: productRepo,
		userGrpcSvc: userGrpcSvc,
	}
}

func (u *usecase) Store(ctx context.Context, product models.Product) (models.Product, error) {
	customCtx, ok := ctx.(restCtx.CustomContext)
	if !ok {
		return models.Product{}, errors.Wrap(errors.New("invalid context"), "product.service.store: parse custom context")
	}

	if customCtx.User()["type"].(string) != models.UserSellerType {
		return models.Product{}, internalErr.NotFoundError("not found")
	}

	product.Seller.ID = customCtx.User()["id"].(string)

	result, err := u.productRepo.Store(ctx, product)
	if err != nil {
		return models.Product{}, errors.Wrap(err, "product.service.store: store from repository")
	}

	return result, nil
}

func (u *usecase) Get(ctx context.Context, ID string) (models.Product, error) {
	result, err := u.productRepo.Get(ctx, ID)
	if err != nil {
		return models.Product{}, errors.Wrap(err, "product.service.get: get from repository")
	}

	users, err := u.userGrpcSvc.FetchByIDs(ctx, []string{result.Seller.ID})
	if err != nil {
		return models.Product{}, errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	result.Seller = users[result.Seller.ID]

	return result, nil
}

func (u *usecase) Fetch(ctx context.Context, filter models.ProductFilter) ([]models.Product, error) {
	customCtx, ok := ctx.(restCtx.CustomContext)
	if ok {
		if customCtx.User()["type"].(string) == models.UserSellerType {
			filter.SellerID = customCtx.User()["id"].(string)
		}
	}

	result, err := u.productRepo.Fetch(ctx, filter)
	if err != nil {
		return make([]models.Product, 0), errors.Wrap(err, "product.service.fetch: fetch from repository")
	}

	userIDs := []string{}
	for _, product := range result {
		userIDs = append(userIDs, product.Seller.ID)
	}

	users, err := u.userGrpcSvc.FetchByIDs(ctx, userIDs)
	if err != nil {
		return make([]models.Product, 0), errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	for index := range result {
		result[index].Seller = users[result[index].Seller.ID]
	}

	return result, nil
}

func (u *usecase) FetchByIds(ctx context.Context, ids []string) (map[string]models.Product, error) {
	result, err := u.productRepo.FetchByIds(ctx, ids)
	if err != nil {
		return map[string]models.Product{}, errors.Wrap(err, "product.service.fetch: fetch from repository")
	}

	userIDs := []string{}
	for _, product := range result {
		userIDs = append(userIDs, product.Seller.ID)
	}

	users, err := u.userGrpcSvc.FetchByIDs(ctx, userIDs)
	if err != nil {
		return map[string]models.Product{}, errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	for index, product := range result {
		product.Seller = users[product.Seller.ID]
		result[index] = product
	}

	return result, nil
}
