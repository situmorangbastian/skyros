package usecase

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/situmorangbastian/skyros/productservice/internal/models"
	"github.com/situmorangbastian/skyros/productservice/internal/repository"
	"github.com/situmorangbastian/skyros/serviceutils/auth"
)

type ProductUsecase interface {
	Store(ctx context.Context, product models.Product) (models.Product, error)
	Get(ctx context.Context, ID string) (models.Product, error)
	Fetch(ctx context.Context, filter models.ProductFilter) ([]models.Product, error)
	FetchByIds(ctx context.Context, ids []string) (map[string]models.Product, error)
}

type usecase struct {
	productRepo repository.ProductRepository
	usrClient   auth.UserClient
}

func NewProductUsecase(productRepo repository.ProductRepository, usrClient auth.UserClient) ProductUsecase {
	return &usecase{
		productRepo: productRepo,
		usrClient:   usrClient,
	}
}

func (u *usecase) Store(ctx context.Context, product models.Product) (models.Product, error) {
	log := zerolog.Ctx(ctx)
	log.With().Str("func", "internal.usecase.product.Store").Logger()

	user, err := auth.GetUserClaims(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed GetUserClaims")
		return models.Product{}, err
	}

	if user.Type != auth.UserSellerType {
		return models.Product{}, status.Error(codes.Unauthenticated, "invalid user")
	}

	product.Seller.ID = user.ID
	result, err := u.productRepo.Store(ctx, product)
	if err != nil {
		log.Error().Err(err).Msg("failed store product")
		return models.Product{}, errors.Wrap(err, "product.service.store: store from repository")
	}

	return result, nil
}

func (u *usecase) Get(ctx context.Context, ID string) (models.Product, error) {
	log := zerolog.Ctx(ctx)
	log.With().Str("func", "internal.usecase.product.Get").Logger()

	result, err := u.productRepo.Get(ctx, ID)
	if err != nil {
		log.Error().Err(err).Msg("failed get product")
		return models.Product{}, errors.Wrap(err, "product.service.get: get from repository")
	}

	users, err := u.usrClient.FetchByIDs(ctx, []string{result.Seller.ID})
	if err != nil {
		log.Error().Err(err).Msg("failed fetch user by ids")
		return models.Product{}, errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	result.Seller = users[result.Seller.ID]
	return result, nil
}

func (u *usecase) Fetch(ctx context.Context, filter models.ProductFilter) ([]models.Product, error) {
	log := zerolog.Ctx(ctx)
	log.With().Str("func", "internal.usecase.product.Fetch").Logger()

	user, err := auth.GetUserClaims(ctx)
	if err == nil {
		if user.Type == auth.UserSellerType {
			filter.SellerID = user.ID
		}
	}

	result, err := u.productRepo.Fetch(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("failed fetch product")
		return make([]models.Product, 0), errors.Wrap(err, "product.service.fetch: fetch from repository")
	}

	userIDs := []string{}
	for _, product := range result {
		userIDs = append(userIDs, product.Seller.ID)
	}

	users, err := u.usrClient.FetchByIDs(ctx, userIDs)
	if err != nil {
		log.Error().Err(err).Msg("failed fetch user by ids")
		return make([]models.Product, 0), errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	for index := range result {
		result[index].Seller = users[result[index].Seller.ID]
	}

	return result, nil
}

func (u *usecase) FetchByIds(ctx context.Context, ids []string) (map[string]models.Product, error) {
	log := zerolog.Ctx(ctx)
	log.With().Str("func", "internal.usecase.product.FetchByIds").Logger()

	result, err := u.productRepo.FetchByIds(ctx, ids)
	if err != nil {
		log.Error().Err(err).Msg("failed fetch product by ids")
		return map[string]models.Product{}, errors.Wrap(err, "product.service.fetch: fetch from repository")
	}

	userIDs := []string{}
	for _, product := range result {
		userIDs = append(userIDs, product.Seller.ID)
	}

	users, err := u.usrClient.FetchByIDs(ctx, userIDs)
	if err != nil {
		log.Error().Err(err).Msg("failed fetch user by ids")
		return map[string]models.Product{}, errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	for index, product := range result {
		product.Seller = users[product.Seller.ID]
		result[index] = product
	}

	return result, nil
}
