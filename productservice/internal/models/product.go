package models

import (
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

type ProductFilter struct {
	Page     int
	PageSize int
	Search   string
	SellerID string
	OrderID  string
}
