package http

import (
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/situmorangbastian/skyros/userservice"
)

func ErrorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			if e, ok := err.(*echo.HTTPError); ok {
				switch e.Code {
				case http.StatusInternalServerError:
					log.Error(e.Message)
					e.Message = "internal server error"
				}

				return echo.NewHTTPError(e.Code, e.Message)
			}

			switch errors.Cause(err).(type) {
			case userservice.ErrorNotFound:
				return echo.NewHTTPError(http.StatusNotFound, errors.Cause(err).Error())
			case userservice.ConflictError:
				return echo.NewHTTPError(http.StatusConflict, errors.Cause(err).Error())
			case userservice.ConstraintError:
				return echo.NewHTTPError(http.StatusBadRequest, errors.Cause(err).Error())
			}

			log.Errorln(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}
	}
}

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
