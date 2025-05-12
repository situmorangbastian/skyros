package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/situmorangbastian/skyros/orderservice/internal/models"
	"github.com/situmorangbastian/skyros/orderservice/internal/usecase"
	"github.com/situmorangbastian/skyros/orderservice/internal/validation"
	orderpb "github.com/situmorangbastian/skyros/proto/order"
	userpb "github.com/situmorangbastian/skyros/proto/user"
)

type service struct {
	usecase   usecase.OrderUsecase
	validator validation.CustomValidator
}

func NewOrderService(usecase usecase.OrderUsecase, validator validation.CustomValidator) orderpb.OrderServiceServer {
	return &service{
		usecase:   usecase,
		validator: validator,
	}
}

func (s *service) CreateOrder(ctx context.Context, request *orderpb.CreateOrderRequest) (*orderpb.Order, error) {
	req := models.Order{
		Description:        request.GetDescription(),
		DestinationAddress: request.GetDestinationAddreess(),
	}

	if request.GetItems() == nil || (request.GetItems() != nil && len(request.GetItems()) == 0) {
		return nil, status.Error(codes.InvalidArgument, "items is required")
	}

	itemsReq := []models.OrderProduct{}
	for _, item := range request.GetItems() {
		itemsReq = append(itemsReq, models.OrderProduct{
			ProductID: item.GetProductId(),
			Quantity:  item.GetQuantity(),
		})
	}
	req.Items = itemsReq

	err := s.validator.Validate(req)
	if err != nil {
		return nil, err
	}

	res, err := s.usecase.Store(ctx, req)
	if err != nil {
		return nil, err
	}

	return &orderpb.Order{
		Id:                  res.ID,
		Description:         res.Description,
		SourceAddress:       res.SourceAddress,
		DestinationAddreess: res.DestinationAddress,
		TotalPrice:          res.TotalPrice,
		Status: func() string {
			status := "pending"
			if res.Status == 1 {
				status = "accepted"
			}
			return status
		}(),
		Seller: &userpb.User{
			Name:    res.Seller.Name,
			Address: res.Seller.Address,
		},
		Buyer: &userpb.User{
			Name:    res.Seller.Name,
			Address: res.Seller.Address,
		},
		Items: func() []*orderpb.OrderProduct {
			items := []*orderpb.OrderProduct{}
			for _, item := range res.Items {
				items = append(items, &orderpb.OrderProduct{
					ProductId: item.ProductID,
					Quantity:  item.Quantity,
				})
			}
			return items
		}(),
		CreatedAt: res.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: res.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *service) GetOrder(ctx context.Context, request *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	res, err := s.usecase.Get(ctx, request.GetOrderId())
	if err != nil {
		return nil, err
	}

	return &orderpb.Order{
		Id:                  res.ID,
		Description:         res.Description,
		SourceAddress:       res.SourceAddress,
		DestinationAddreess: res.DestinationAddress,
		TotalPrice:          res.TotalPrice,
		Status: func() string {
			status := "pending"
			if res.Status == 1 {
				status = "accepted"
			}
			return status
		}(),
		Seller: &userpb.User{
			Name:    res.Seller.Name,
			Address: res.Seller.Address,
		},
		Buyer: &userpb.User{
			Name:    res.Seller.Name,
			Address: res.Seller.Address,
		},
		Items: func() []*orderpb.OrderProduct {
			items := []*orderpb.OrderProduct{}
			for _, item := range res.Items {
				items = append(items, &orderpb.OrderProduct{
					ProductId: item.ProductID,
					Quantity:  item.Quantity,
				})
			}
			return items
		}(),
		CreatedAt: res.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: res.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *service) GetOrders(ctx context.Context, request *orderpb.GetOrdersRequest) (*orderpb.GetOrdersResponse, error) {
	limit := request.GetLimit()
	if limit == 0 {
		limit = 20
	}

	orders, err := s.usecase.Fetch(ctx, models.Filter{
		PageSize: int(limit),
		Page:     int(request.GetOffset()),
		Search:   request.GetSearch(),
	})
	if err != nil {
		return nil, err
	}

	result := []*orderpb.Order{}
	for _, order := range orders {
		result = append(result, &orderpb.Order{
			Id:                  order.ID,
			Description:         order.Description,
			SourceAddress:       order.SourceAddress,
			DestinationAddreess: order.DestinationAddress,
			TotalPrice:          order.TotalPrice,
			Status: func() string {
				status := "pending"
				if order.Status == 1 {
					status = "accepted"
				}
				return status
			}(),
			Seller: &userpb.User{
				Name:    order.Seller.Name,
				Address: order.Seller.Address,
			},
			Buyer: &userpb.User{
				Name:    order.Seller.Name,
				Address: order.Seller.Address,
			},
			Items: func() []*orderpb.OrderProduct {
				items := []*orderpb.OrderProduct{}
				for _, item := range order.Items {
					items = append(items, &orderpb.OrderProduct{
						ProductId: item.ProductID,
						Quantity:  item.Quantity,
					})
				}
				return items
			}(),
			CreatedAt: order.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: order.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &orderpb.GetOrdersResponse{
		Result: result,
	}, nil
}
