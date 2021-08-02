package product

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/situmorangbastian/skyros"
)

type service struct {
	productRepo skyros.ProductRepository
	userRepo    skyros.UserRepository
}

func NewService(productRepo skyros.ProductRepository, userRepo skyros.UserRepository) skyros.ProductService {
	return service{
		productRepo: productRepo,
		userRepo:    userRepo,
	}
}

func (s service) Store(ctx context.Context, product skyros.Product) (skyros.Product, error) {
	customCtx, ok := ctx.(skyros.CustomContext)
	if !ok {
		return skyros.Product{}, errors.Wrap(errors.New("invalid context"), "product.service.store: parse custom context")
	}

	if customCtx.User().Type != skyros.UserSellerType {
		return skyros.Product{}, skyros.ErrorNotFound("not found")
	}

	product.Seller = customCtx.User()

	result, err := s.productRepo.Store(ctx, product)
	if err != nil {
		return skyros.Product{}, errors.Wrap(err, "product.service.store: store from repository")
	}

	return result, nil
}

func (s service) Get(ctx context.Context, ID string) (skyros.Product, error) {
	result, err := s.productRepo.Get(ctx, ID)
	if err != nil {
		return skyros.Product{}, errors.Wrap(err, "product.service.get: get from repository")
	}

	detailSeller, err := s.userRepo.GetUser(ctx, result.Seller.ID)
	if err != nil {
		return skyros.Product{}, errors.Wrap(err, "product.service.get: get user from repository")
	}

	result.Seller = detailSeller

	return result, nil
}

func (s service) Fetch(ctx context.Context, filter skyros.Filter) ([]skyros.Product, string, error) {
	customCtx, ok := ctx.(skyros.CustomContext)
	if ok {
		if customCtx.User().Type == skyros.UserSellerType {
			filter.SellerID = customCtx.User().ID
		}
	}

	result, nextCursor, err := s.productRepo.Fetch(ctx, filter)
	if err != nil {
		return make([]skyros.Product, 0), "", errors.Wrap(err, "product.service.fetch: fetch from repository")
	}

	errGroup := errgroup.Group{}
	for index, product := range result {
		index, product := index, product

		errGroup.Go(func() error {
			detailSeller, err := s.userRepo.GetUser(ctx, product.Seller.ID)
			if err != nil {
				return errors.Wrap(err, "product.service.fetch: get user from repository")
			}

			result[index].Seller = detailSeller
			return nil
		})

	}

	if err := errGroup.Wait(); err != nil {
		return []skyros.Product{}, "", errors.Wrap(err, "resolve seller detail on product")
	}

	return result, nextCursor, nil
}
