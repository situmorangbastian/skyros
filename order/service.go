package order

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/situmorangbastian/skyros"
)

type service struct {
	orderRepo   skyros.OrderRepository
	productRepo skyros.ProductRepository
}

func NewService(orderRepo skyros.OrderRepository, productRepo skyros.ProductRepository) skyros.OrderService {
	return service{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s service) Store(ctx context.Context, order skyros.Order) (skyros.Order, error) {
	customCtx, ok := ctx.(skyros.CustomContext)
	if !ok {
		return skyros.Order{}, errors.Wrap(errors.New("invalid context"), "order.service.store: parse custom context")
	}

	if customCtx.User().Type != skyros.UserBuyerType {
		return skyros.Order{}, skyros.ErrorNotFound("not found")
	}

	order.Buyer = customCtx.User()

	errGroup := errgroup.Group{}

	order.TotalPrice = 0
	for index, orderItem := range order.Items {
		index, orderItem := index, orderItem

		errGroup.Go(func() error {
			productDetail, err := s.productRepo.Get(ctx, orderItem.ProductID)
			if err != nil {
				return errors.Wrap(err, "get detail product id: "+orderItem.ProductID)
			}

			order.Items[index].Product = productDetail
			order.TotalPrice += order.Items[index].Product.Price * order.Items[index].Quantity
			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return skyros.Order{}, errors.Wrap(err, "resolve product detail on order item")
	}

	result, err := s.orderRepo.Store(ctx, order)
	if err != nil {
		return skyros.Order{}, errors.Wrap(err, "order.service.store: store from repository")
	}

	return result, nil
}

func (s service) Get(ctx context.Context, ID string) (skyros.Order, error) {
	customCtx, ok := ctx.(skyros.CustomContext)
	if !ok {
		return skyros.Order{}, errors.Wrap(errors.New("invalid context"), "order.service.get: parse custom context")
	}

	filter := skyros.Filter{
		OrderID: ID,
	}

	switch customCtx.User().Type {
	case skyros.UserBuyerType:
		filter.BuyerID = customCtx.User().ID
	case skyros.UserSellerType:
		filter.SellerID = customCtx.User().ID
	default:
		return skyros.Order{}, skyros.ErrorNotFound("not found")
	}

	result, _, err := s.orderRepo.Fetch(ctx, filter)
	if err != nil {
		return skyros.Order{}, errors.Wrap(err, "order.service.get: fetch from repository")
	}

	if len(result) == 0 {
		return skyros.Order{}, skyros.ErrorNotFound("not found")
	}

	return result[0], nil
}

func (s service) Fetch(ctx context.Context, filter skyros.Filter) ([]skyros.Order, string, error) {
	customCtx, ok := ctx.(skyros.CustomContext)
	if !ok {
		return []skyros.Order{}, "", errors.Wrap(errors.New("invalid context"), "order.service.fetch: parse custom context")
	}

	switch customCtx.User().Type {
	case skyros.UserBuyerType:
		filter.BuyerID = customCtx.User().ID
	case skyros.UserSellerType:
		filter.SellerID = customCtx.User().ID
	default:
		return []skyros.Order{}, "", skyros.ErrorNotFound("not found")
	}

	result, cursor, err := s.orderRepo.Fetch(ctx, filter)
	if err != nil {
		return []skyros.Order{}, "", errors.Wrap(err, "order.service.fetch: fetch from repository")
	}

	return result, cursor, nil
}

func (s service) Accept(ctx context.Context, ID string) error {
	customCtx, ok := ctx.(skyros.CustomContext)
	if !ok {
		return errors.Wrap(errors.New("invalid context"), "order.service.accept: parse custom context")
	}

	if customCtx.User().Type != skyros.UserSellerType {
		return skyros.ErrorNotFound("not found")
	}

	return s.orderRepo.Accept(ctx, ID)
}