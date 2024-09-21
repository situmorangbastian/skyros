// internal/usecase/order_usecase.go
package usecase

import (
	"context"

	"github.com/pkg/errors"

	restCtx "github.com/situmorangbastian/skyros/orderservice/api/rest/context"
	"github.com/situmorangbastian/skyros/orderservice/internal/domain/models"
	internalErr "github.com/situmorangbastian/skyros/orderservice/internal/errors"
	"github.com/situmorangbastian/skyros/orderservice/internal/repository"
	"github.com/situmorangbastian/skyros/orderservice/internal/services"
)

type OrderUsecase interface {
	Store(ctx context.Context, order models.Order) (models.Order, error)
	Get(ctx context.Context, ID string) (models.Order, error)
	Fetch(ctx context.Context, filter models.Filter) ([]models.Order, error)
	PatchStatus(ctx context.Context, ID string, status int) error
}

type usecase struct {
	orderRepo      repository.OrderRepository
	userService    services.UserServiceGrpc
	productService services.ProductServiceGrpc
}

func NewUsecase(
	orderRepo repository.OrderRepository,
	userService services.UserServiceGrpc,
	productService services.ProductServiceGrpc) OrderUsecase {
	return &usecase{
		orderRepo:      orderRepo,
		userService:    userService,
		productService: productService,
	}
}

func (u *usecase) Store(ctx context.Context, order models.Order) (models.Order, error) {
	customCtx, ok := ctx.(restCtx.CustomContext)
	if !ok {
		return models.Order{}, errors.Wrap(errors.New("invalid context"), "order.service.store: parse custom context")
	}

	if customCtx.User()["type"].(string) != models.UserBuyerType {
		return models.Order{}, internalErr.NotFoundError("not found")
	}

	order.Buyer.ID = customCtx.User()["id"].(string)

	order.TotalPrice = 0
	productIds := []string{}
	for _, item := range order.Items {
		productIds = append(productIds, item.Product.ID)
	}

	products, err := u.productService.FetchByIDs(ctx, productIds)
	if err != nil {
		return models.Order{}, errors.Wrap(err, "order.service.store: fetch product")
	}

	for index := range order.Items {
		order.Items[index].Product = products[order.Items[index].Product.ID]
		if order.Items[index].Product.Name == "" {
			return models.Order{}, errors.Wrap(internalErr.NotFoundError("product not found"),
				"order.service.store: fetch product")
		}
		order.Seller = order.Items[index].Product.Seller
		order.TotalPrice += order.Items[index].Product.Price * order.Items[index].Quantity
	}

	result, err := u.orderRepo.Store(ctx, order)
	if err != nil {
		return models.Order{}, errors.Wrap(err, "order.service.store: store from repository")
	}

	return result, nil
}

func (u *usecase) Get(ctx context.Context, ID string) (models.Order, error) {
	customCtx, ok := ctx.(restCtx.CustomContext)
	if !ok {
		return models.Order{}, errors.Wrap(errors.New("invalid context"), "order.service.get: parse custom context")
	}

	filter := models.Filter{
		OrderID: ID,
	}

	switch customCtx.User()["type"].(string) {
	case models.UserBuyerType:
		filter.BuyerID = customCtx.User()["id"].(string)
	case models.UserSellerType:
		filter.SellerID = customCtx.User()["id"].(string)
	default:
		return models.Order{}, internalErr.NotFoundError("not found")
	}

	result, err := u.orderRepo.Fetch(ctx, filter)
	if err != nil {
		return models.Order{}, errors.Wrap(err, "order.service.get: fetch from repository")
	}

	if len(result) == 0 {
		return models.Order{}, internalErr.NotFoundError("not found")
	}

	users, err := u.userService.FetchByIDs(ctx, []string{result[0].Seller.ID, result[0].Buyer.ID})
	if err != nil {
		return models.Order{}, errors.Wrap(err, "order.service.get: fetch users")
	}

	result[0].Buyer = users[result[0].Buyer.ID]
	result[0].Seller = users[result[0].Seller.ID]

	productIds := []string{}
	for _, order := range result {
		for _, item := range order.Items {
			productIds = append(productIds, item.ProductID)
		}
	}

	products, err := u.productService.FetchByIDs(ctx, productIds)
	if err != nil {
		return models.Order{}, errors.Wrap(err, "order.service.get: fetch product")
	}

	for index, order := range result {
		for index := range order.Items {
			order.Items[index].Product = products[order.Items[index].ProductID]
		}
		result[index].Items = order.Items
	}

	return result[0], nil
}

func (u *usecase) Fetch(ctx context.Context, filter models.Filter) ([]models.Order, error) {
	customCtx, ok := ctx.(restCtx.CustomContext)
	if !ok {
		return []models.Order{}, errors.Wrap(errors.New("invalid context"), "order.service.fetch: parse custom context")
	}

	switch customCtx.User()["type"].(string) {
	case models.UserBuyerType:
		filter.BuyerID = customCtx.User()["id"].(string)
	case models.UserSellerType:
		filter.SellerID = customCtx.User()["id"].(string)
	default:
		return []models.Order{}, internalErr.NotFoundError("not found")
	}

	result, err := u.orderRepo.Fetch(ctx, filter)
	if err != nil {
		return []models.Order{}, errors.Wrap(err, "order.service.fetch: fetch from repository")
	}

	userIds := []string{}
	productIds := []string{}
	for _, order := range result {
		userIds = append(userIds, order.Buyer.ID, order.Seller.ID)
		for _, item := range order.Items {
			productIds = append(productIds, item.ProductID)
		}
	}

	users, err := u.userService.FetchByIDs(ctx, userIds)
	if err != nil {
		return []models.Order{}, errors.Wrap(err, "order.service.fetch: fetch users")
	}

	products, err := u.productService.FetchByIDs(ctx, productIds)
	if err != nil {
		return []models.Order{}, errors.Wrap(err, "order.service.fetch: fetch products")
	}

	for index, order := range result {
		result[index].Seller = users[result[index].Seller.ID]
		result[index].Buyer = users[result[index].Buyer.ID]
		for index := range order.Items {
			order.Items[index].Product = products[order.Items[index].ProductID]
		}
		result[index].Items = order.Items
	}

	return result, nil
}

func (u *usecase) PatchStatus(ctx context.Context, ID string, status int) error {
	customCtx, ok := ctx.(restCtx.CustomContext)
	if !ok {
		return errors.Wrap(errors.New("invalid context"), "order.service.accept: parse custom context")
	}

	if customCtx.User()["type"].(string) != models.UserSellerType {
		return internalErr.NotFoundError("not found")
	}

	return u.orderRepo.PatchStatus(ctx, ID, status)
}
