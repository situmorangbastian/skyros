package order

import (
	"context"

	"github.com/pkg/errors"

	"github.com/situmorangbastian/eclipse"
	"github.com/situmorangbastian/skyros/orderservice"
)

type service struct {
	orderRepo      orderservice.OrderRepository
	userService    orderservice.UserServiceGrpc
	productService orderservice.ProductServiceGrpc
}

func NewService(
	orderRepo orderservice.OrderRepository,
	userService orderservice.UserServiceGrpc,
	productService orderservice.ProductServiceGrpc) orderservice.OrderService {
	return service{
		orderRepo:      orderRepo,
		userService:    userService,
		productService: productService,
	}
}

func (s service) Store(ctx context.Context, order orderservice.Order) (orderservice.Order, error) {
	customCtx, ok := ctx.(eclipse.CustomContext)
	if !ok {
		return orderservice.Order{}, errors.Wrap(errors.New("invalid context"), "order.service.store: parse custom context")
	}

	if customCtx.User()["type"].(string) != orderservice.UserBuyerType {
		return orderservice.Order{}, eclipse.NotFoundError("not found")
	}

	order.Buyer.ID = customCtx.User()["id"].(string)

	order.TotalPrice = 0
	productIds := []string{}
	for _, item := range order.Items {
		productIds = append(productIds, item.Product.ID)
	}

	products, err := s.productService.FetchByIDs(ctx, productIds)
	if err != nil {
		return orderservice.Order{}, errors.Wrap(err, "order.service.store: fetch product")
	}

	for index := range order.Items {
		order.Items[index].Product = products[order.Items[index].Product.ID]
		if order.Items[index].Product.Name == "" {
			return orderservice.Order{}, errors.Wrap(eclipse.NotFoundError("product not found"),
				"order.service.store: fetch product")
		}
		order.Seller = order.Items[index].Product.Seller
		order.TotalPrice += order.Items[index].Product.Price * order.Items[index].Quantity
	}

	result, err := s.orderRepo.Store(ctx, order)
	if err != nil {
		return orderservice.Order{}, errors.Wrap(err, "order.service.store: store from repository")
	}

	return result, nil
}

func (s service) Get(ctx context.Context, ID string) (orderservice.Order, error) {
	customCtx, ok := ctx.(eclipse.CustomContext)
	if !ok {
		return orderservice.Order{}, errors.Wrap(errors.New("invalid context"), "order.service.get: parse custom context")
	}

	filter := orderservice.Filter{
		OrderID: ID,
	}

	switch customCtx.User()["type"].(string) {
	case orderservice.UserBuyerType:
		filter.BuyerID = customCtx.User()["id"].(string)
	case orderservice.UserSellerType:
		filter.SellerID = customCtx.User()["id"].(string)
	default:
		return orderservice.Order{}, eclipse.NotFoundError("not found")
	}

	result, _, err := s.orderRepo.Fetch(ctx, filter)
	if err != nil {
		return orderservice.Order{}, errors.Wrap(err, "order.service.get: fetch from repository")
	}

	if len(result) == 0 {
		return orderservice.Order{}, eclipse.NotFoundError("not found")
	}

	users, err := s.userService.FetchByIDs(ctx, []string{result[0].Seller.ID, result[0].Buyer.ID})
	if err != nil {
		return orderservice.Order{}, errors.Wrap(err, "order.service.get: fetch users")
	}

	result[0].Buyer = users[result[0].Buyer.ID]
	result[0].Seller = users[result[0].Seller.ID]

	productIds := []string{}
	for _, order := range result {
		for _, item := range order.Items {
			productIds = append(productIds, item.ProductID)
		}
	}

	products, err := s.productService.FetchByIDs(ctx, productIds)
	if err != nil {
		return orderservice.Order{}, errors.Wrap(err, "order.service.get: fetch product")
	}

	for index, order := range result {
		for index := range order.Items {
			order.Items[index].Product = products[order.Items[index].ProductID]
		}
		result[index].Items = order.Items
	}

	return result[0], nil
}

func (s service) Fetch(ctx context.Context, filter orderservice.Filter) ([]orderservice.Order, string, error) {
	customCtx, ok := ctx.(eclipse.CustomContext)
	if !ok {
		return []orderservice.Order{}, "", errors.Wrap(errors.New("invalid context"), "order.service.fetch: parse custom context")
	}

	switch customCtx.User()["type"].(string) {
	case orderservice.UserBuyerType:
		filter.BuyerID = customCtx.User()["id"].(string)
	case orderservice.UserSellerType:
		filter.SellerID = customCtx.User()["id"].(string)
	default:
		return []orderservice.Order{}, "", eclipse.NotFoundError("not found")
	}

	result, cursor, err := s.orderRepo.Fetch(ctx, filter)
	if err != nil {
		return []orderservice.Order{}, "", errors.Wrap(err, "order.service.fetch: fetch from repository")
	}

	userIds := []string{}
	productIds := []string{}
	for _, order := range result {
		userIds = append(userIds, order.Buyer.ID, order.Seller.ID)
		for _, item := range order.Items {
			productIds = append(productIds, item.ProductID)
		}
	}

	users, err := s.userService.FetchByIDs(ctx, userIds)
	if err != nil {
		return []orderservice.Order{}, "", errors.Wrap(err, "order.service.fetch: fetch users")
	}

	products, err := s.productService.FetchByIDs(ctx, productIds)
	if err != nil {
		return []orderservice.Order{}, "", errors.Wrap(err, "order.service.fetch: fetch products")
	}

	for index, order := range result {
		result[index].Seller = users[result[index].Seller.ID]
		result[index].Buyer = users[result[index].Buyer.ID]
		for index := range order.Items {
			order.Items[index].Product = products[order.Items[index].ProductID]
		}
		result[index].Items = order.Items
	}

	return result, cursor, nil
}

func (s service) PatchStatus(ctx context.Context, ID string, status int) error {
	customCtx, ok := ctx.(eclipse.CustomContext)
	if !ok {
		return errors.Wrap(errors.New("invalid context"), "order.service.accept: parse custom context")
	}

	if customCtx.User()["type"].(string) != orderservice.UserSellerType {
		return eclipse.NotFoundError("not found")
	}

	return s.orderRepo.PatchStatus(ctx, ID, status)
}
