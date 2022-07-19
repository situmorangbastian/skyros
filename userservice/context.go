package userservice

import (
	"context"
)

type CustomContext struct {
	context.Context
	user User
}

func (c CustomContext) User() User {
	return c.user
}

func NewCustomContext(c context.Context, u User) CustomContext {
	return CustomContext{Context: c, user: u}
}
