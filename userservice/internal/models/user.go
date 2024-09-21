package models

import (
	"encoding/json"
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
