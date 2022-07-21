package http

import (
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

	"github.com/situmorangbastian/skyros/userservice"
)

func Authentication() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "token invalid/expired/required")
			}
			claims := user.Claims.(jwt.MapClaims)
			id := claims["id"].(string)
			address := claims["address"].(string)
			name := claims["name"].(string)
			email := claims["email"].(string)
			type_ := claims["type"].(string)

			validUser := userservice.User{
				ID:      id,
				Address: address,
				Name:    name,
				Email:   email,
				Type:    type_,
			}

			c.SetRequest(c.Request().WithContext(userservice.NewCustomContext(c.Request().Context(), validUser)))

			return next(c)
		}
	}
}
