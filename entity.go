package skyros

import (
	"encoding/json"
	"time"
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
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	Seller      User      `json:"seller"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
}

type Filter struct {
	Cursor   string
	Num      int
	Search   string
	SellerID string
}
