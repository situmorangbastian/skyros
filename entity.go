package skyros

import (
	"encoding/json"
	"time"
)

const (
	UserSellerType = "seller"
	UserBuyerType  = "buyer"
)

type User struct {
	ID       string `json:"-"`
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
	Address  string `json:"address" validate:"required"`
	Password string `json:"password" validate:"required"`
	Type     string `json:"-"`
}

func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Address string `json:"address"`
		Type    string `json:"type"`
	}{
		ID:      u.ID,
		Email:   u.Email,
		Name:    u.Name,
		Address: u.Address,
		Type:    u.Type,
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

type Order struct {
	ID                 string         `json:"id"`
	Buyer              User           `json:"buyer"`
	Seller             User           `json:"seller"`
	Description        string         `json:"description"`
	SourceAddress      string         `json:"source_address"`
	DestinationAddress string         `json:"destination_address"`
	Items              []OrderProduct `json:"items"`
	TotalPrice         int64          `json:"total_price"`
	Status             int            `json:"status"`
}

type OrderProduct struct {
	Product  Product `json:"product"`
	Quantity int64   `json:"quantity"`
}

type Filter struct {
	Cursor   string
	Num      int
	Search   string
	SellerID string
	BuyerID  string
	OrderID  string
}
