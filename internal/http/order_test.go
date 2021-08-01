package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/situmorangbastian/skyros"
	"github.com/situmorangbastian/skyros/internal"
	handler "github.com/situmorangbastian/skyros/internal/http"
	"github.com/situmorangbastian/skyros/mocks"
	"github.com/situmorangbastian/skyros/testdata"
)

func TestOrderHTTP_Store(t *testing.T) {
	var mockOrder skyros.Order
	testdata.GoldenJSONUnmarshal(t, "order", &mockOrder)

	tests := []struct {
		testName       string
		orderService   testdata.FuncCall
		input          []byte
		expectedStatus int
	}{
		{
			testName: "success",
			orderService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockOrder},
				Output: []interface{}{mockOrder, nil},
			},
			input:          testdata.GetGolden(t, "order"),
			expectedStatus: http.StatusCreated,
		},
		{
			testName:       "error invalid request body",
			input:          []byte(`{"description": "medan",}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "unexpected error from service",
			orderService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockOrder},
				Output: []interface{}{skyros.Order{}, errors.New("unexpected error")},
			},
			input:          testdata.GetGolden(t, "order"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	e := echo.New()
	e.Use(handler.ErrorMiddleware())
	e.Validator = internal.NewValidator()

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockOrderService := new(mocks.OrderService)
			if test.orderService.Called {
				mockOrderService.On("Store", test.orderService.Input...).
					Return(test.orderService.Output...).Once()
			}
			handler.NewOrderHandler(e, mockOrderService)

			req := httptest.NewRequest(echo.POST, "/order", strings.NewReader(string(test.input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			mockOrderService.AssertExpectations(t)
			require.Equal(t, test.expectedStatus, rec.Code)
		})
	}
}

func TestOrderHTTP_Get(t *testing.T) {
	var mockOrder skyros.Order
	testdata.GoldenJSONUnmarshal(t, "order", &mockOrder)

	tests := []struct {
		testName       string
		orderService   testdata.FuncCall
		expectedStatus int
	}{
		{
			testName: "success",
			orderService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockOrder.ID},
				Output: []interface{}{mockOrder, nil},
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName: "unexpected error from service",
			orderService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockOrder.ID},
				Output: []interface{}{skyros.Order{}, errors.New("unexpected error")},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	e := echo.New()
	e.Use(handler.ErrorMiddleware())
	e.Validator = internal.NewValidator()

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockOrderService := new(mocks.OrderService)
			if test.orderService.Called {
				mockOrderService.On("Get", test.orderService.Input...).
					Return(test.orderService.Output...).Once()
			}
			handler.NewOrderHandler(e, mockOrderService)

			req := httptest.NewRequest(echo.GET, "/order/"+mockOrder.ID, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			mockOrderService.AssertExpectations(t)
			require.Equal(t, test.expectedStatus, rec.Code)
		})
	}
}

func TestOrderHTTP_Fetch(t *testing.T) {
	var mockOrder skyros.Order
	testdata.GoldenJSONUnmarshal(t, "order", &mockOrder)

	tests := []struct {
		testName              string
		requestURL            string
		orderService          testdata.FuncCall
		expectedStatus        int
		expectedHeaderXCursor string
	}{
		{
			testName:   "success",
			requestURL: "/order?num=20&cursor=next-cursor&search=mobil",
			orderService: testdata.FuncCall{
				Called: true,
				Input: []interface{}{mock.Anything, skyros.Filter{
					Num:    20,
					Search: "mobil",
					Cursor: "next-cursor",
				}},
				Output: []interface{}{[]skyros.Order{mockOrder}, "next-cursor", nil},
			},
			expectedStatus:        http.StatusOK,
			expectedHeaderXCursor: "next-cursor",
		},
		{
			testName:   "unexpected error from service",
			requestURL: "/order?num=20&cursor=next-cursor&search=mobil",
			orderService: testdata.FuncCall{
				Called: true,
				Input: []interface{}{mock.Anything, skyros.Filter{
					Num:    20,
					Search: "mobil",
					Cursor: "next-cursor",
				}},
				Output: []interface{}{[]skyros.Order{}, "", errors.New("unexpected error")},
			},
			expectedStatus:        http.StatusInternalServerError,
			expectedHeaderXCursor: "",
		},
		{
			testName:              "error invalid request query param num",
			requestURL:            "/order?num=abc",
			expectedStatus:        http.StatusBadRequest,
			expectedHeaderXCursor: "",
		},
	}

	e := echo.New()
	e.Use(handler.ErrorMiddleware())
	e.Validator = internal.NewValidator()

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockOrderService := new(mocks.OrderService)
			if test.orderService.Called {
				mockOrderService.On("Fetch", test.orderService.Input...).
					Return(test.orderService.Output...).Once()
			}
			handler.NewOrderHandler(e, mockOrderService)

			req := httptest.NewRequest(echo.GET, test.requestURL, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			mockOrderService.AssertExpectations(t)
			require.Equal(t, test.expectedStatus, rec.Code)
			require.Equal(t, test.expectedHeaderXCursor, rec.Header().Get("X-Cursor"))
		})
	}
}

func TestOrderHTTP_PatchStatus(t *testing.T) {
	tests := []struct {
		testName       string
		orderService   testdata.FuncCall
		input          []byte
		expectedStatus int
	}{
		{
			testName: "success",
			orderService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "order-id"},
				Output: []interface{}{nil},
			},
			input:          []byte(`{"status": "accept"}`),
			expectedStatus: http.StatusNoContent,
		},
		{
			testName:       "error invalid status",
			input:          []byte(`{"status": "other-status"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "unexpected error from service",
			orderService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "order-id"},
				Output: []interface{}{errors.New("unexpected error")},
			},
			input:          []byte(`{"status": "accept"}`),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	e := echo.New()
	e.Use(handler.ErrorMiddleware())
	e.Validator = internal.NewValidator()

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockOrderService := new(mocks.OrderService)
			if test.orderService.Called {
				mockOrderService.On("Accept", test.orderService.Input...).
					Return(test.orderService.Output...).Once()
			}
			handler.NewOrderHandler(e, mockOrderService)

			req := httptest.NewRequest(echo.PATCH, "/order/order-id", strings.NewReader(string(test.input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			mockOrderService.AssertExpectations(t)
			require.Equal(t, test.expectedStatus, rec.Code)
		})
	}
}
