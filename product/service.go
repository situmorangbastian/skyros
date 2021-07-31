package product

import (
	"context"

	"github.com/pkg/errors"

	"github.com/situmorangbastian/skyros"
)

type service struct {
	repo skyros.ProductRepository
}

func NewService(repo skyros.ProductRepository) skyros.ProductService {
	return service{
		repo: repo,
	}
}

func (s service) Store(ctx context.Context, product skyros.Product) (skyros.Product, error) {
	customCtx, ok := ctx.(skyros.CustomContext)
	if !ok {
		return skyros.Product{}, errors.Wrap(errors.New("invalid context"), "product.service.store: parse custom context")
	}
	product.Seller = customCtx.User()

	result, err := s.repo.Store(ctx, product)
	if err != nil {
		return skyros.Product{}, errors.Wrap(err, "product.service.store: store from repository")
	}

	return result, nil
}

func (s service) Get(ctx context.Context, ID string) (skyros.Product, error) {
	result, err := s.repo.Get(ctx, ID)
	if err != nil {
		return skyros.Product{}, errors.Wrap(err, "product.service.get: get from repository")
	}

	return result, nil
}

func (s service) Fetch(ctx context.Context, filter skyros.Filter) ([]skyros.Product, string, error) {
	customCtx, ok := ctx.(skyros.CustomContext)
	if ok {
		if customCtx.User().Type == "seller" {
			filter.SellerID = customCtx.User().ID
		}
	}

	result, nextCursor, err := s.repo.Fetch(ctx, filter)
	if err != nil {
		return make([]skyros.Product, 0), "", errors.Wrap(err, "product.service.fetch: fetch from repository")
	}

	return result, nextCursor, nil
}
