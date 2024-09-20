package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	resthandlers "github.com/situmorangbastian/skyros/orderservice/api/rest/handlers"
	"github.com/situmorangbastian/skyros/orderservice/api/rest/validators"
	"github.com/situmorangbastian/skyros/orderservice/internal/domain/models"
	internalErr "github.com/situmorangbastian/skyros/orderservice/internal/error"
	"github.com/situmorangbastian/skyros/orderservice/mocks"
	"github.com/situmorangbastian/skyros/orderservice/testdata"
)

func getEchoServer() *echo.Echo {
	e := echo.New()
	e.Use(internalErr.Error())
	e.Validator = validators.NewValidator()

	return e
}

func TestOrderHTTPStore(t *testing.T) {
	orderJSON := testdata.GetGolden(t, "order")
	invalidOrderJSON := testdata.GetGolden(t, "invalidorder")

	var mockOrder models.Order
	testdata.GoldenJSONUnmarshal(t, "order", &mockOrder)

	tests := map[string]struct {
		input              []byte
		orderUsecase       testdata.FuncCall
		expectedStatusCode int
	}{
		"success": {
			input: orderJSON,
			orderUsecase: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockOrder},
				Output: []interface{}{mockOrder, nil},
			},
			expectedStatusCode: http.StatusCreated,
		},
		"error: invalid request body": {
			input:              []byte(`invalid request body`),
			expectedStatusCode: http.StatusBadRequest,
		},
		"error: validating destination address": {
			input:              invalidOrderJSON,
			expectedStatusCode: http.StatusBadRequest,
		},
		"error: unexpected error from service": {
			input: orderJSON,
			orderUsecase: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockOrder},
				Output: []interface{}{models.Order{}, errors.New("unexpected error")},
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	e := getEchoServer()
	g := e.Group("")
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			service := new(mocks.OrderUsecase)
			if test.orderUsecase.Called {
				service.On("Store", test.orderUsecase.Input...).
					Return(test.orderUsecase.Output...).Once()
			}

			req := httptest.NewRequest(echo.POST, "/order", strings.NewReader(string(test.input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			resthandlers.NewOrderHandler(g, service)
			e.ServeHTTP(rec, req)

			service.AssertExpectations(t)

			require.Equal(t, test.expectedStatusCode, rec.Code)
		})
	}
}

func TestOrderHTTPGet(t *testing.T) {
	var mockOrder models.Order
	testdata.GoldenJSONUnmarshal(t, "order", &mockOrder)

	tests := map[string]struct {
		orderId            string
		orderUsecase       testdata.FuncCall
		expectedStatusCode int
	}{
		"success": {
			orderId: "order-id",
			orderUsecase: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "order-id"},
				Output: []interface{}{mockOrder, nil},
			},
			expectedStatusCode: http.StatusOK,
		},
		"error: unexpected error from service": {
			orderId: "order-id",
			orderUsecase: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "order-id"},
				Output: []interface{}{models.Order{}, errors.New("unexpected error")},
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	e := getEchoServer()
	g := e.Group("")
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			service := new(mocks.OrderUsecase)
			if test.orderUsecase.Called {
				service.On("Get", test.orderUsecase.Input...).
					Return(test.orderUsecase.Output...).Once()
			}

			req := httptest.NewRequest(echo.GET, "/order/"+test.orderId, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			resthandlers.NewOrderHandler(g, service)
			e.ServeHTTP(rec, req)

			service.AssertExpectations(t)

			require.Equal(t, test.expectedStatusCode, rec.Code)
		})
	}
}

func TestOrderHTTPFetch(t *testing.T) {
	var mockOrders []models.Order
	testdata.GoldenJSONUnmarshal(t, "orders", &mockOrders)

	tests := map[string]struct {
		queryParam         string
		orderUsecase       testdata.FuncCall
		expectedStatusCode int
	}{
		"success": {
			queryParam: "?pagesize=20",
			orderUsecase: testdata.FuncCall{
				Called: true,
				Input: []interface{}{mock.Anything, models.Filter{
					PageSize: 20,
				}},
				Output: []interface{}{mockOrders, nil},
			},
			expectedStatusCode: http.StatusOK,
		},
		"error: unexpected error from service": {
			queryParam: "?pagesize=20&page=1",
			orderUsecase: testdata.FuncCall{
				Called: true,
				Input: []interface{}{mock.Anything, models.Filter{
					PageSize: 20,
					Page:     1,
				}},
				Output: []interface{}{[]models.Order{}, errors.New("unexpected error")},
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		"error: invalid query param limit": {
			queryParam:         "?pagesize=abc",
			expectedStatusCode: http.StatusBadRequest,
		},
		"error: invalid query param offset": {
			queryParam:         "?page=abc",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	e := getEchoServer()
	g := e.Group("")
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			service := new(mocks.OrderUsecase)
			if test.orderUsecase.Called {
				service.On("Fetch", test.orderUsecase.Input...).
					Return(test.orderUsecase.Output...).Once()
			}

			req := httptest.NewRequest(echo.GET, "/order"+test.queryParam, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			resthandlers.NewOrderHandler(g, service)
			e.ServeHTTP(rec, req)

			service.AssertExpectations(t)

			require.Equal(t, test.expectedStatusCode, rec.Code)
		})
	}
}

func TestOrderHTTPPatch(t *testing.T) {
	tests := map[string]struct {
		input              []byte
		orderUsecase       testdata.FuncCall
		expectedStatusCode int
	}{
		"success": {
			input: []byte(`{"status":"accept"}`),
			orderUsecase: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "order-id", 1},
				Output: []interface{}{nil},
			},
			expectedStatusCode: http.StatusNoContent,
		},
		"error: invalid request body": {
			input:              []byte(`invalid request body`),
			expectedStatusCode: http.StatusBadRequest,
		},
		"error: validating status": {
			input:              []byte(`{"status":""}`),
			expectedStatusCode: http.StatusBadRequest,
		},
		"error: validating status value": {
			input:              []byte(`{"status":"denied"}`),
			expectedStatusCode: http.StatusBadRequest,
		},
		"error: unexpected error from service": {
			input: []byte(`{"status":"accept"}`),
			orderUsecase: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "order-id", 1},
				Output: []interface{}{errors.New("unexpected error")},
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	e := getEchoServer()
	g := e.Group("")
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			service := new(mocks.OrderUsecase)
			if test.orderUsecase.Called {
				service.On("PatchStatus", test.orderUsecase.Input...).
					Return(test.orderUsecase.Output...).Once()
			}

			req := httptest.NewRequest(echo.PATCH, "/order/order-id", strings.NewReader(string(test.input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			resthandlers.NewOrderHandler(g, service)
			e.ServeHTTP(rec, req)

			service.AssertExpectations(t)

			require.Equal(t, test.expectedStatusCode, rec.Code)
		})
	}
}
