package models

import "time"

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Price       int64     `json:"price" validate:"required"`
	Seller      User      `json:"seller" validate:"-"`
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
