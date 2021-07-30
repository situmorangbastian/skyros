package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/situmorangbastian/skyros"
)

// ErrorMiddleware is a function to generate http status code.
func ErrorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			if e, ok := err.(*echo.HTTPError); ok {
				switch e.Code {
				case http.StatusUnauthorized:
					e.Message = "token invalid/expired"
				case http.StatusInternalServerError:
					log.Error(e.Message)
					e.Message = "internal server error"
				}

				return echo.NewHTTPError(e.Code, e.Message)
			}

			switch errors.Cause(err).(type) {
			case skyros.ErrorNotFound:
				return echo.NewHTTPError(http.StatusNotFound, errors.Cause(err).Error())
			case skyros.ConflictError:
				return echo.NewHTTPError(http.StatusConflict, errors.Cause(err).Error())
			case skyros.ConstraintError:
				return echo.NewHTTPError(http.StatusBadRequest, errors.Cause(err).Error())
			}

			log.Errorln(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}
	}
}
