package models

import "encoding/json"

const (
	UserSellerType = "seller"
	UserBuyerType  = "buyer"
)

type User struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Address string `json:"address"`
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
