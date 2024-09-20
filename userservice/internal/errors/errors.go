package errors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type ConstraintError string

func (e ConstraintError) Error() string {
	return string(e)
}

func ConstraintErrorf(format string, a ...interface{}) ConstraintError {
	return ConstraintError(fmt.Sprintf(format, a...))
}

type NotFoundError string

func (e NotFoundError) Error() string {
	return string(e)
}

func NotFoundErrorf(format string, a ...interface{}) NotFoundError {
	return NotFoundError(fmt.Sprintf(format, a...))
}

type ConflictError string

func (e ConflictError) Error() string {
	return string(e)
}

func ConflictErrorf(format string, a ...interface{}) ConflictError {
	return ConflictError(fmt.Sprintf(format, a...))
}

func Error() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			if e, ok := err.(*echo.HTTPError); ok {
				if e.Code >= http.StatusInternalServerError {
					log.Error(e.Message)
				}

				return echo.NewHTTPError(e.Code, strings.ToLower(e.Message.(string)))
			}

			// Check error based on error type
			switch errors.Cause(err).(type) {
			case ConstraintError:
				return echo.NewHTTPError(http.StatusBadRequest, errors.Cause(err).Error())
			case NotFoundError:
				return echo.NewHTTPError(http.StatusNotFound, errors.Cause(err).Error())
			case ConflictError:
				return echo.NewHTTPError(http.StatusConflict, errors.Cause(err).Error())
			default:
				log.Error(err)
				return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
			}
		}
	}
}
