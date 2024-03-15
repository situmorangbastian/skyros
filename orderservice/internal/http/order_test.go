package http_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/situmorangbastian/skyros/orderservice"
	"github.com/situmorangbastian/skyros/orderservice/internal"
	handler "github.com/situmorangbastian/skyros/orderservice/internal/http"
	"github.com/situmorangbastian/skyros/orderservice/mocks"
	"github.com/situmorangbastian/skyros/orderservice/testdata"
)

func getEchoServer() *echo.Echo {
	e := echo.New()
	e.Use(orderservice.Error())
	e.Validator = internal.NewValidator()

	return e
}

func TestOrderHTTPStore(t *testing.T) {
	orderJSON := testdata.GetGolden(t, "order")
	invalidOrderJSON := testdata.GetGolden(t, "invalidorder")

	var mockOrder orderservice.Order
	testdata.GoldenJSONUnmarshal(t, "order", &mockOrder)

	tests := map[string]struct {
		input              []byte
		orderService       testdata.FuncCall
		expectedStatusCode int
	}{
		"success": {
			input: orderJSON,
			orderService: testdata.FuncCall{
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
			orderService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockOrder},
				Output: []interface{}{orderservice.Order{}, errors.New("unexpected error")},
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	e := getEchoServer()
	g := e.Group("")
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			service := new(mocks.OrderService)
			if test.orderService.Called {
				service.On("Store", test.orderService.Input...).
					Return(test.orderService.Output...).Once()
			}

			req := httptest.NewRequest(echo.POST, "/order", strings.NewReader(string(test.input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			handler.NewOrderHandler(g, service)
			e.ServeHTTP(rec, req)

			service.AssertExpectations(t)

			require.Equal(t, test.expectedStatusCode, rec.Code)
		})
	}
}

func TestOrderHTTPGet(t *testing.T) {
	var mockOrder orderservice.Order
	testdata.GoldenJSONUnmarshal(t, "order", &mockOrder)

	tests := map[string]struct {
		orderId            string
		orderService       testdata.FuncCall
		expectedStatusCode int
	}{
		"success": {
			orderId: "order-id",
			orderService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "order-id"},
				Output: []interface{}{mockOrder, nil},
			},
			expectedStatusCode: http.StatusOK,
		},
		"error: unexpected error from service": {
			orderId: "order-id",
			orderService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "order-id"},
				Output: []interface{}{orderservice.Order{}, errors.New("unexpected error")},
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	e := getEchoServer()
	g := e.Group("")
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			service := new(mocks.OrderService)
			if test.orderService.Called {
				service.On("Get", test.orderService.Input...).
					Return(test.orderService.Output...).Once()
			}

			req := httptest.NewRequest(echo.GET, "/order/"+test.orderId, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			handler.NewOrderHandler(g, service)
			e.ServeHTTP(rec, req)

			service.AssertExpectations(t)

			require.Equal(t, test.expectedStatusCode, rec.Code)
		})
	}
}

func TestOrderHTTPFetch(t *testing.T) {
	var mockOrders []orderservice.Order
	testdata.GoldenJSONUnmarshal(t, "orders", &mockOrders)

	tests := map[string]struct {
		queryParam         string
		orderService       testdata.FuncCall
		expectedStatusCode int
	}{
		"success": {
			queryParam: "?num=20",
			orderService: testdata.FuncCall{
				Called: true,
				Input: []interface{}{mock.Anything, orderservice.Filter{
					Num: 20,
				}},
				Output: []interface{}{mockOrders, "", nil},
			},
			expectedStatusCode: http.StatusOK,
		},
		"error: unexpected error from service": {
			queryParam: "?num=20",
			orderService: testdata.FuncCall{
				Called: true,
				Input: []interface{}{mock.Anything, orderservice.Filter{
					Num: 20,
				}},
				Output: []interface{}{[]orderservice.Order{}, "", errors.New("unexpected error")},
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		"error: invalid query param num": {
			queryParam:         "?num=abc",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	e := getEchoServer()
	g := e.Group("")
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			service := new(mocks.OrderService)
			if test.orderService.Called {
				service.On("Fetch", test.orderService.Input...).
					Return(test.orderService.Output...).Once()
			}

			req := httptest.NewRequest(echo.GET, "/order"+test.queryParam, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			handler.NewOrderHandler(g, service)
			e.ServeHTTP(rec, req)

			service.AssertExpectations(t)

			require.Equal(t, test.expectedStatusCode, rec.Code)
		})
	}
}

func TestOrderHTTPPatch(t *testing.T) {
	tests := map[string]struct {
		input              []byte
		orderService       testdata.FuncCall
		expectedStatusCode int
	}{
		"success": {
			input: []byte(`{"status":"accept"}`),
			orderService: testdata.FuncCall{
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
			orderService: testdata.FuncCall{
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
			service := new(mocks.OrderService)
			if test.orderService.Called {
				service.On("PatchStatus", test.orderService.Input...).
					Return(test.orderService.Output...).Once()
			}

			req := httptest.NewRequest(echo.PATCH, "/order/order-id", strings.NewReader(string(test.input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			handler.NewOrderHandler(g, service)
			e.ServeHTTP(rec, req)

			service.AssertExpectations(t)

			require.Equal(t, test.expectedStatusCode, rec.Code)
		})
	}
}
