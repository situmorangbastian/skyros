package order

import (
	"context"

	"github.com/pkg/errors"
	"github.com/situmorangbastian/skyros/orderservice"
)

type service struct {
	orderRepo orderservice.OrderRepository
}

func NewService(orderRepo orderservice.OrderRepository) orderservice.OrderService {
	return service{
		orderRepo: orderRepo,
	}
}

func (s service) Store(ctx context.Context, order orderservice.Order) (orderservice.Order, error) {
	customCtx, ok := ctx.(orderservice.CustomContext)
	if !ok {
		return orderservice.Order{}, errors.Wrap(errors.New("invalid context"), "order.service.store: parse custom context")
	}

	if customCtx.User().Type != orderservice.UserBuyerType {
		return orderservice.Order{}, orderservice.ErrorNotFound("not found")
	}

	order.Buyer = customCtx.User()

	// errGroup := errgroup.Group{}

	order.TotalPrice = 0
	// for index, orderItem := range order.Items {
	// 	index, orderItem := index, orderItem

	// 	errGroup.Go(func() error {
	// 		productDetail, err := s.productService.Get(ctx, orderItem.ProductID)
	// 		if err != nil {
	// 			return errors.Wrap(err, "get detail product id: "+orderItem.ProductID)
	// 		}

	// 		order.Items[index].Product = productDetail
	// 		order.Seller = productDetail.Seller
	// 		order.TotalPrice += order.Items[index].Product.Price * order.Items[index].Quantity
	// 		return nil
	// 	})
	// }

	// if err := errGroup.Wait(); err != nil {
	// 	return orderservice.Order{}, errors.Wrap(err, "resolve product detail on order item")
	// }

	result, err := s.orderRepo.Store(ctx, order)
	if err != nil {
		return orderservice.Order{}, errors.Wrap(err, "order.service.store: store from repository")
	}

	return result, nil
}

func (s service) Get(ctx context.Context, ID string) (orderservice.Order, error) {
	customCtx, ok := ctx.(orderservice.CustomContext)
	if !ok {
		return orderservice.Order{}, errors.Wrap(errors.New("invalid context"), "order.service.get: parse custom context")
	}

	filter := orderservice.Filter{
		OrderID: ID,
	}

	switch customCtx.User().Type {
	case orderservice.UserBuyerType:
		filter.BuyerID = customCtx.User().ID
	case orderservice.UserSellerType:
		filter.SellerID = customCtx.User().ID
	default:
		return orderservice.Order{}, orderservice.ErrorNotFound("not found")
	}

	result, _, err := s.orderRepo.Fetch(ctx, filter)
	if err != nil {
		return orderservice.Order{}, errors.Wrap(err, "order.service.get: fetch from repository")
	}

	if len(result) == 0 {
		return orderservice.Order{}, orderservice.ErrorNotFound("not found")
	}

	// detailBuyer, err := s.userRepo.GetUser(ctx, result[0].Buyer.ID)
	// if err != nil {
	// 	return orderservice.Order{}, errors.Wrap(err, "order.service.get: resolve detail buyer")
	// }

	// result[0].Buyer = detailBuyer

	// detailSeller, err := s.userRepo.GetUser(ctx, result[0].Seller.ID)
	// if err != nil {
	// 	return orderservice.Order{}, errors.Wrap(err, "order.service.get: resolve detail seller")
	// }

	// result[0].Seller = detailSeller

	// errGroup := errgroup.Group{}

	// for index, orderItem := range result[0].Items {
	// 	index, orderItem := index, orderItem

	// 	errGroup.Go(func() error {
	// 		productDetail, err := s.productService.Get(ctx, orderItem.ProductID)
	// 		if err != nil {
	// 			return errors.Wrap(err, "get detail product id: "+orderItem.ProductID)
	// 		}

	// 		result[0].Items[index].Product = productDetail
	// 		return nil
	// 	})
	// }

	// if err := errGroup.Wait(); err != nil {
	// 	return orderservice.Order{}, errors.Wrap(err, "resolve product detail on order item")
	// }

	return result[0], nil
}

func (s service) Fetch(ctx context.Context, filter orderservice.Filter) ([]orderservice.Order, string, error) {
	customCtx, ok := ctx.(orderservice.CustomContext)
	if !ok {
		return []orderservice.Order{}, "", errors.Wrap(errors.New("invalid context"), "order.service.fetch: parse custom context")
	}

	switch customCtx.User().Type {
	case orderservice.UserBuyerType:
		filter.BuyerID = customCtx.User().ID
	case orderservice.UserSellerType:
		filter.SellerID = customCtx.User().ID
	default:
		return []orderservice.Order{}, "", orderservice.ErrorNotFound("not found")
	}

	result, cursor, err := s.orderRepo.Fetch(ctx, filter)
	if err != nil {
		return []orderservice.Order{}, "", errors.Wrap(err, "order.service.fetch: fetch from repository")
	}

	// errGroup := errgroup.Group{}

	// for index, order := range result {
	// 	index, order := index, order

	// 	errGroup.Go(func() error {
	// 		detailBuyer, err := s.userRepo.GetUser(ctx, order.Buyer.ID)
	// 		if err != nil {
	// 			return errors.Wrap(err, "order.service.fetch: resolve detail buyer")
	// 		}

	// 		result[0].Buyer = detailBuyer

	// 		detailSeller, err := s.userRepo.GetUser(ctx, order.Seller.ID)
	// 		if err != nil {
	// 			return errors.Wrap(err, "order.service.fetch: resolve detail buyer")
	// 		}

	// 		result[0].Seller = detailSeller
	// 		return nil
	// 	})

	// 	errGroupChild := errgroup.Group{}

	// 	for index, orderItem := range result[index].Items {
	// 		index, orderItem := index, orderItem

	// 		errGroupChild.Go(func() error {
	// 			productDetail, err := s.productService.Get(ctx, orderItem.ProductID)
	// 			if err != nil {
	// 				return errors.Wrap(err, "get detail product id: "+orderItem.ProductID)
	// 			}

	// 			result[index].Items[index].Product = productDetail
	// 			return nil
	// 		})
	// 	}

	// 	if err := errGroupChild.Wait(); err != nil {
	// 		return []orderservice.Order{}, "", errors.Wrap(err, "resolve product detail on order item")
	// 	}
	// }

	// if err := errGroup.Wait(); err != nil {
	// 	return []orderservice.Order{}, "", errors.Wrap(err, "resolve seller and buyer detail")
	// }

	return result, cursor, nil
}

func (s service) PatchStatus(ctx context.Context, ID string, status int) error {
	customCtx, ok := ctx.(orderservice.CustomContext)
	if !ok {
		return errors.Wrap(errors.New("invalid context"), "order.service.accept: parse custom context")
	}

	if customCtx.User().Type != orderservice.UserSellerType {
		return orderservice.ErrorNotFound("not found")
	}

	return s.orderRepo.PatchStatus(ctx, ID, status)
}
