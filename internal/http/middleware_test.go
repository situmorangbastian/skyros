package http_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/situmorangbastian/skyros"
	handler "github.com/situmorangbastian/skyros/internal/http"
)

func TestErrorMiddleware(t *testing.T) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)

	mw := handler.ErrorMiddleware()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e := echo.New()
	c := e.NewContext(req, rec)

	t.Run("with not found error", func(t *testing.T) {
		h := func(c echo.Context) error {
			return skyros.ErrorNotFoundf("not found")
		}

		err := mw(h)(c).(*echo.HTTPError)

		require.Error(t, err)
		require.Equal(t, http.StatusNotFound, err.Code)
		require.Contains(t, err.Error(), "not found")
	})

	t.Run("with constraint error", func(t *testing.T) {
		h := func(c echo.Context) error {
			return skyros.ConstraintErrorf("this is a constraint error")
		}

		err := mw(h)(c).(*echo.HTTPError)
		require.Error(t, err)
		require.Equal(t, http.StatusBadRequest, err.Code)
		require.Contains(t, err.Error(), "this is a constraint error")
	})

	t.Run("with unknown error", func(t *testing.T) {
		h := func(c echo.Context) error {
			return errors.New("unexpected error")
		}

		buf := new(bytes.Buffer)
		log.SetOutput(buf)

		err := mw(h)(c).(*echo.HTTPError)
		require.Error(t, err)
		require.Equal(t, http.StatusInternalServerError, err.Code)
		require.Contains(t, buf.String(), "unexpected error")
	})

	t.Run("with unexpected error from echo", func(t *testing.T) {
		h := func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusInternalServerError, "unexpected error")
		}

		buf := new(bytes.Buffer)
		log.SetOutput(buf)

		err := mw(h)(c).(*echo.HTTPError)
		require.Error(t, err)
		require.Equal(t, http.StatusInternalServerError, err.Code)
		require.Contains(t, buf.String(), "unexpected error")
	})

	t.Run("with unauthorized error from echo", func(t *testing.T) {
		h := func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusUnauthorized, "token invalid/expired")
		}

		buf := new(bytes.Buffer)
		log.SetOutput(buf)

		err := mw(h)(c).(*echo.HTTPError)
		require.Error(t, err)
		require.Equal(t, http.StatusUnauthorized, err.Code)
		require.Equal(t, "token invalid/expired", err.Message.(string))
	})
}
