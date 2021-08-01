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
	Buyer              User           `json:"buyer" validate:"-"`
	Seller             User           `json:"seller" validate:"-"`
	Description        string         `json:"description"`
	SourceAddress      string         `json:"source_address"`
	DestinationAddress string         `json:"destination_address" validate:"required"`
	Items              []OrderProduct `json:"items" validate:"required,min=1"`
	TotalPrice         int64          `json:"total_price"`
	Status             int            `json:"status"`
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
	Cursor   string
	Num      int
	Search   string
	SellerID string
	BuyerID  string
	OrderID  string
}
