package orderservice

import (
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

func Authentication() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "token invalid/expired/required")
			}
			userClaims := user.Claims.(jwt.MapClaims)

			c.SetRequest(c.Request().WithContext(NewCustomContext(c.Request().Context(), userClaims)))

			return next(c)
		}
	}
}
