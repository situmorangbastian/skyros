package models

import (
	"encoding/json"
	"time"

	"github.com/situmorangbastian/skyros/serviceutils/auth"
)

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Price       int64     `json:"price" validate:"required"`
	Seller      auth.User `json:"seller" validate:"-"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
}
type Order struct {
	ID                 string         `json:"id"`
	Buyer              auth.User      `json:"buyer" validate:"-"`
	Seller             auth.User      `json:"seller" validate:"-"`
	Description        string         `json:"description"`
	SourceAddress      string         `json:"source_address"`
	DestinationAddress string         `json:"destination_address" validate:"required"`
	Items              []OrderProduct `json:"items" validate:"required,min=1"`
	TotalPrice         int64          `json:"total_price"`
	Status             int            `json:"status"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

func (o Order) MarshalJSON() ([]byte, error) {
	status := "pending"
	if o.Status == 1 {
		status = "accepted"
	}

	return json.Marshal(&struct {
		ID                 string         `json:"id"`
		Buyer              auth.User      `json:"buyer" validate:"-"`
		Seller             auth.User      `json:"seller" validate:"-"`
		Description        string         `json:"description"`
		SourceAddress      string         `json:"source_address"`
		DestinationAddress string         `json:"destination_address" validate:"required"`
		Items              []OrderProduct `json:"items" validate:"required,min=1"`
		TotalPrice         int64          `json:"total_price"`
		Status             string         `json:"status"`
		CreatedAt          time.Time      `json:"created_at"`
		UpdatedAt          time.Time      `json:"updated_at"`
	}{
		ID:                 o.ID,
		Buyer:              o.Buyer,
		Seller:             o.Seller,
		Description:        o.Description,
		SourceAddress:      o.SourceAddress,
		DestinationAddress: o.DestinationAddress,
		Items:              o.Items,
		TotalPrice:         o.TotalPrice,
		Status:             status,
		CreatedAt:          o.CreatedAt,
		UpdatedAt:          o.UpdatedAt,
	})
}

type OrderProduct struct {
	Product   Product `json:"-" validate:"-"`
	ProductID string  `json:"product_id" validate:"required"`
	Quantity  int64   `json:"quantity" validate:"min=1"`
}

func (op OrderProduct) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Product  Product `json:"product"`
		Quantity int64   `json:"quantity"`
	}{
		Product:  op.Product,
		Quantity: op.Quantity,
	})
}

type Filter struct {
	Page     int
	PageSize int
	Search   string
	SellerID string
	BuyerID  string
	OrderID  string
}
