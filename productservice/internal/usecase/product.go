package usecase

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/situmorangbastian/skyros/productservice/internal/integration"
	"github.com/situmorangbastian/skyros/productservice/internal/models"
	"github.com/situmorangbastian/skyros/productservice/internal/repository"
	"github.com/situmorangbastian/skyros/productservice/middleware"
)

type ProductUsecase interface {
	Store(ctx context.Context, product models.Product) (models.Product, error)
	Get(ctx context.Context, ID string) (models.Product, error)
	Fetch(ctx context.Context, filter models.ProductFilter) ([]models.Product, error)
	FetchByIds(ctx context.Context, ids []string) (map[string]models.Product, error)
}

type usecase struct {
	productRepo repository.ProductRepository
	usrClient   integration.UserClient
}

func NewProductUsecase(productRepo repository.ProductRepository, usrClient integration.UserClient) ProductUsecase {
	return &usecase{
		productRepo: productRepo,
		usrClient:   usrClient,
	}
}

func (u *usecase) Store(ctx context.Context, product models.Product) (models.Product, error) {
	claims, ok := middleware.GetUserClaims(ctx)
	if !ok {
		return models.Product{}, status.Error(codes.Unauthenticated, "failed get user claims")
	}

	if claims["type"].(string) != models.UserSellerType {
		return models.Product{}, status.Error(codes.NotFound, "Not Found")
	}

	product.Seller.ID = claims["id"].(string)

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

	users, err := u.usrClient.FetchByIDs(ctx, []string{result.Seller.ID})
	if err != nil {
		return models.Product{}, errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	result.Seller = users[result.Seller.ID]

	return result, nil
}

func (u *usecase) Fetch(ctx context.Context, filter models.ProductFilter) ([]models.Product, error) {
	claims, ok := middleware.GetUserClaims(ctx)
	if ok {
		userType := claims["type"].(string)
		if userType == models.UserSellerType {
			filter.SellerID = claims["id"].(string)
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

	users, err := u.usrClient.FetchByIDs(ctx, userIDs)
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

	users, err := u.usrClient.FetchByIDs(ctx, userIDs)
	if err != nil {
		return map[string]models.Product{}, errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	for index, product := range result {
		product.Seller = users[product.Seller.ID]
		result[index] = product
	}

	return result, nil
}
