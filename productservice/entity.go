package productservice

import (
	"encoding/json"
	"time"
)

const (
	UserSellerType = "seller"
	UserBuyerType  = "buyer"
)

type User struct {
	ID      string `json:"id"`
	Email   string `json:"email" validate:"required,email"`
	Name    string `json:"name" validate:"required"`
	Address string `json:"address" validate:"required"`
	Type    string `json:"type"`
}

func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Address string `json:"address"`
	}{
		Email:   u.Email,
		Name:    u.Name,
		Address: u.Address,
	})
}

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Price       int64     `json:"price" validate:"required"`
	Seller      User      `json:"seller" validate:"-"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
}

type Filter struct {
	Cursor   string
	Num      int
	Search   string
	SellerID string
	OrderID  string
}
