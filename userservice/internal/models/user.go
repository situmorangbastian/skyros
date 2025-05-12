package models

import (
	"encoding/json"
)

type User struct {
	ID       string   `json:"-"`
	Email    string   `json:"email" validate:"required,email"`
	Name     string   `json:"name" validate:"required"`
	Data     UserData `json:"data"`
	Password string   `json:"password" validate:"required"`
}

type UserData struct {
	Address string `json:"address"`
	Type    string `json:"type"`
}

func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
	})
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
