// internal/usecase/order_usecase.go
package usecase

import (
	"context"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/situmorangbastian/skyros/orderservice/internal/integration"
	"github.com/situmorangbastian/skyros/orderservice/internal/models"
	"github.com/situmorangbastian/skyros/orderservice/internal/repository"
	"github.com/situmorangbastian/skyros/serviceutils/auth"
)

type OrderUsecase interface {
	Store(ctx context.Context, order models.Order) (models.Order, error)
	Get(ctx context.Context, ID string) (models.Order, error)
	Fetch(ctx context.Context, filter models.Filter) ([]models.Order, error)
	PatchStatus(ctx context.Context, ID string, status int) error
}

type usecase struct {
	orderRepo     repository.OrderRepository
	userClient    auth.UserClient
	productClient integration.ProductClient
	logger        zerolog.Logger
}

func NewUsecase(
	orderRepo repository.OrderRepository,
	userClient auth.UserClient,
	productClient integration.ProductClient,
	logger zerolog.Logger) OrderUsecase {
	return &usecase{
		orderRepo:     orderRepo,
		userClient:    userClient,
		productClient: productClient,
		logger:        logger,
	}
}

func (u *usecase) Store(ctx context.Context, order models.Order) (models.Order, error) {
	log := u.logger.With().Str("func", "internal.usecase.user.Store").Logger()

	user, err := auth.GetUserClaims(ctx)
	if err != nil {
		return models.Order{}, err
	}

	if user.Type != auth.UserBuyerType {
		return models.Order{}, status.Error(codes.NotFound, "Not Found")
	}

	order.Buyer.ID = user.ID
	order.TotalPrice = 0
	productIds := []string{}
	for _, item := range order.Items {
		productIds = append(productIds, item.ProductID)
	}

	products, err := u.productClient.FetchByIDs(ctx, productIds)
	if err != nil {
		log.Error().Err(err).Msg("failed FetchByIDs")
		return models.Order{}, status.Error(codes.Internal, "Internal Server Error")
	}

	for index := range order.Items {
		order.Items[index].Product = products[order.Items[index].ProductID]
		if order.Items[index].Product.Name == "" {
			return models.Order{}, status.Error(codes.NotFound, "product not found")
		}
		order.Seller = order.Items[index].Product.Seller
		order.TotalPrice += order.Items[index].Product.Price * order.Items[index].Quantity
	}

	result, err := u.orderRepo.Store(ctx, order)
	if err != nil {
		log.Error().Err(err).Msg("failed Store")
		return models.Order{}, status.Error(codes.Internal, "Internal Server Error")
	}

	return result, nil
}

func (u *usecase) Get(ctx context.Context, ID string) (models.Order, error) {
	log := u.logger.With().Str("func", "internal.usecase.user.Get").Logger()

	user, err := auth.GetUserClaims(ctx)
	if err != nil {
		return models.Order{}, err
	}

	filter := models.Filter{
		OrderID:  ID,
		PageSize: 20,
	}

	switch user.Type {
	case auth.UserBuyerType:
		filter.BuyerID = user.ID
	case auth.UserSellerType:
		filter.SellerID = user.ID
	default:
		return models.Order{}, status.Error(codes.NotFound, "Not Found")
	}

	result, err := u.orderRepo.Fetch(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("failed Fetch")
		return models.Order{}, status.Error(codes.Internal, "Internal Server Error")
	}

	if len(result) == 0 {
		return models.Order{}, status.Error(codes.NotFound, "Not Found")
	}

	users, err := u.userClient.FetchByIDs(ctx, []string{result[0].Seller.ID, result[0].Buyer.ID})
	if err != nil {
		log.Error().Err(err).Msg("failed FetchByIDs")
		return models.Order{}, status.Error(codes.Internal, "Internal Server Error")
	}

	result[0].Buyer = users[result[0].Buyer.ID]
	result[0].Seller = users[result[0].Seller.ID]

	productIds := []string{}
	for _, order := range result {
		for _, item := range order.Items {
			productIds = append(productIds, item.ProductID)
		}
	}

	products, err := u.productClient.FetchByIDs(ctx, productIds)
	if err != nil {
		log.Error().Err(err).Msg("failed FetchByIDs")
		return models.Order{}, status.Error(codes.Internal, "Internal Server Error")
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
	log := u.logger.With().Str("func", "internal.usecase.user.Fetch").Logger()

	user, err := auth.GetUserClaims(ctx)
	if err != nil {
		return []models.Order{}, err
	}

	switch user.Type {
	case auth.UserBuyerType:
		filter.BuyerID = user.ID
	case auth.UserSellerType:
		filter.SellerID = user.ID
	default:
		return []models.Order{}, status.Error(codes.NotFound, "Not Found")
	}

	result, err := u.orderRepo.Fetch(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("failed Fetch")
		return []models.Order{}, status.Error(codes.Internal, "Internal Server Error")
	}

	userIds := []string{}
	productIds := []string{}
	for _, order := range result {
		userIds = append(userIds, order.Buyer.ID, order.Seller.ID)
		for _, item := range order.Items {
			productIds = append(productIds, item.ProductID)
		}
	}

	users, err := u.userClient.FetchByIDs(ctx, userIds)
	if err != nil {
		log.Error().Err(err).Msg("failed FetchByIDs")
		return []models.Order{}, status.Error(codes.Internal, "Internal Server Error")
	}

	products, err := u.productClient.FetchByIDs(ctx, productIds)
	if err != nil {
		log.Error().Err(err).Msg("failed FetchByIDs")
		return []models.Order{}, status.Error(codes.Internal, "Internal Server Error")
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

func (u *usecase) PatchStatus(ctx context.Context, ID string, statusOrder int) error {
	log := u.logger.With().Str("func", "internal.usecase.user.PatchStatus").Logger()

	user, err := auth.GetUserClaims(ctx)
	if err != nil {
		return err
	}

	if user.Type != auth.UserSellerType {
		return status.Error(codes.NotFound, "Not Found")
	}

	err = u.orderRepo.PatchStatus(ctx, ID, statusOrder)
	if err != nil {
		log.Error().Err(err).Msg("failed PatchStatus")
		return status.Error(codes.Internal, "Internal Server Error")
	}

	return nil
}
