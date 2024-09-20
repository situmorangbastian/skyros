package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

	"github.com/situmorangbastian/skyros/productservice/api/rest/context"
)

func Authentication() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*jwt.Token)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "token invalid/expired/required")
			}
			userClaims := user.Claims.(jwt.MapClaims)

			c.SetRequest(c.Request().WithContext(context.NewCustomContext(c.Request().Context(), userClaims)))

			return next(c)
		}
	}
}
