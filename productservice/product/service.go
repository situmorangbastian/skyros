package product

import (
	"context"

	"github.com/pkg/errors"
	"github.com/situmorangbastian/eclipse"

	"github.com/situmorangbastian/skyros/productservice"
)

type service struct {
	productRepo     productservice.ProductRepository
	userServiceGrpc productservice.UserServiceGrpc
}

func NewService(productRepo productservice.ProductRepository, userServiceGrpc productservice.UserServiceGrpc) productservice.ProductService {
	return service{
		productRepo:     productRepo,
		userServiceGrpc: userServiceGrpc,
	}
}

func (s service) Store(ctx context.Context, product productservice.Product) (productservice.Product, error) {
	customCtx, ok := ctx.(eclipse.CustomContext)
	if !ok {
		return productservice.Product{}, errors.Wrap(errors.New("invalid context"), "product.service.store: parse custom context")
	}

	if customCtx.User()["type"].(string) != productservice.UserSellerType {
		return productservice.Product{}, eclipse.NotFoundError("not found")
	}

	product.Seller.ID = customCtx.User()["id"].(string)

	result, err := s.productRepo.Store(ctx, product)
	if err != nil {
		return productservice.Product{}, errors.Wrap(err, "product.service.store: store from repository")
	}

	return result, nil
}

func (s service) Get(ctx context.Context, ID string) (productservice.Product, error) {
	result, err := s.productRepo.Get(ctx, ID)
	if err != nil {
		return productservice.Product{}, errors.Wrap(err, "product.service.get: get from repository")
	}

	users, err := s.userServiceGrpc.FetchByIDs(ctx, []string{result.Seller.ID})
	if err != nil {
		return productservice.Product{}, errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	result.Seller = users[result.Seller.ID]

	return result, nil
}

func (s service) Fetch(ctx context.Context, filter productservice.Filter) ([]productservice.Product, string, error) {
	customCtx, ok := ctx.(eclipse.CustomContext)
	if ok {
		if customCtx.User()["type"].(string) == productservice.UserSellerType {
			filter.SellerID = customCtx.User()["id"].(string)
		}
	}

	result, nextCursor, err := s.productRepo.Fetch(ctx, filter)
	if err != nil {
		return make([]productservice.Product, 0), "", errors.Wrap(err, "product.service.fetch: fetch from repository")
	}

	userIDs := []string{}
	for _, product := range result {
		userIDs = append(userIDs, product.Seller.ID)
	}

	users, err := s.userServiceGrpc.FetchByIDs(ctx, userIDs)
	if err != nil {
		return make([]productservice.Product, 0), "", errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	for index := range result {
		result[index].Seller = users[result[index].Seller.ID]
	}

	return result, nextCursor, nil
}

func (s service) FetchByIds(ctx context.Context, ids []string) (map[string]productservice.Product, error) {
	result, err := s.productRepo.FetchByIds(ctx, ids)
	if err != nil {
		return map[string]productservice.Product{}, errors.Wrap(err, "product.service.fetch: fetch from repository")
	}

	userIDs := []string{}
	for _, product := range result {
		userIDs = append(userIDs, product.Seller.ID)
	}

	users, err := s.userServiceGrpc.FetchByIDs(ctx, userIDs)
	if err != nil {
		return map[string]productservice.Product{}, errors.Wrap(err, "product.service.get: get user from userservice grpc")
	}

	for index, product := range result {
		product.Seller = users[product.Seller.ID]
		result[index] = product
	}

	return result, nil
}
