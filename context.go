package skyros

import (
	"context"
)

// CustomContext represents extended context of the current HTTP request.
type CustomContext struct {
	context.Context
	user User
}

// User active on current request
func (c CustomContext) User() User {
	return c.user
}

// NewCustomContext is a constructor for custom context.
func NewCustomContext(c context.Context, u User) CustomContext {
	return CustomContext{Context: c, user: u}
}
