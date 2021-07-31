package http_test

import (
	"encoding/json"
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

func TestProductHTTP_Store(t *testing.T) {
	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(t, "product", &mockProduct)

	tests := []struct {
		testName       string
		productService testdata.FuncCall
		input          []byte
		expectedStatus int
	}{
		{
			testName: "success",
			productService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockProduct},
				Output: []interface{}{mockProduct, nil},
			},
			input:          testdata.GetGolden(t, "product"),
			expectedStatus: http.StatusCreated,
		},
		{
			testName: "unexpected error from service",
			productService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockProduct},
				Output: []interface{}{skyros.Product{}, errors.New("unexpected error")},
			},
			input:          testdata.GetGolden(t, "product"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			testName:       "invalid body request",
			input:          []byte(`{"name": "minyak",}`),
			expectedStatus: http.StatusBadRequest,
		},
	}

	e := echo.New()
	e.Use(handler.ErrorMiddleware())
	e.Validator = internal.NewValidator()

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockProductService := new(mocks.ProductService)
			if test.productService.Called {
				mockProductService.On("Store", test.productService.Input...).
					Return(test.productService.Output...).Once()
			}
			handler.NewProductHandler(e, mockProductService)

			req := httptest.NewRequest(echo.POST, "/product", strings.NewReader(string(test.input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			mockProductService.AssertExpectations(t)
			require.Equal(t, test.expectedStatus, rec.Code)

			if test.expectedStatus == http.StatusCreated {
				var result skyros.Product
				err := json.Unmarshal(rec.Body.Bytes(), &result)
				require.NoError(t, err)
				require.Equal(t, mockProduct, result)
			}
		})
	}
}

func TestProductHTTP_Get(t *testing.T) {
	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(t, "product", &mockProduct)

	tests := []struct {
		testName       string
		productService testdata.FuncCall
		expectedStatus int
	}{
		{
			testName: "success",
			productService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockProduct.ID},
				Output: []interface{}{mockProduct, nil},
			},
			expectedStatus: http.StatusOK,
		},
		{
			testName: "unexpected error from service",
			productService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockProduct.ID},
				Output: []interface{}{skyros.Product{}, errors.New("unexpected error")},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	e := echo.New()
	e.Use(handler.ErrorMiddleware())

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockProductService := new(mocks.ProductService)
			if test.productService.Called {
				mockProductService.On("Get", test.productService.Input...).
					Return(test.productService.Output...).Once()
			}
			handler.NewProductHandler(e, mockProductService)

			req := httptest.NewRequest(echo.GET, "/product/"+mockProduct.ID, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			mockProductService.AssertExpectations(t)

			require.Equal(t, test.expectedStatus, rec.Code)
		})
	}
}

func TestProductHTTP_Fetch(t *testing.T) {
	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(t, "product", &mockProduct)

	mockProducts := []skyros.Product{mockProduct}

	passedFilter := skyros.Filter{
		Num:    20,
		Cursor: "next-cursor",
		Search: "minyak",
	}

	tests := []struct {
		testName       string
		productService testdata.FuncCall
		expectedTarget string
		expectedCursor string
		expectedStatus int
	}{
		{
			testName: "success",
			productService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, passedFilter},
				Output: []interface{}{mockProducts, "next-cursor", nil},
			},
			expectedTarget: "/product?num=20&cursor=next-cursor&search=minyak",
			expectedCursor: "next-cursor",
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "invalid num",
			expectedTarget: "/product?num=abc",
			expectedCursor: "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "unexpected error from service",
			productService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, passedFilter},
				Output: []interface{}{[]skyros.Product{}, "", errors.New("unexpected error")},
			},
			expectedTarget: "/product?num=20&cursor=next-cursor&search=minyak",
			expectedCursor: "",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	e := echo.New()
	e.Use(handler.ErrorMiddleware())

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockProductService := new(mocks.ProductService)
			if test.productService.Called {
				mockProductService.On("Fetch", test.productService.Input...).
					Return(test.productService.Output...).Once()
			}
			handler.NewProductHandler(e, mockProductService)

			req := httptest.NewRequest(echo.GET, test.expectedTarget, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			mockProductService.AssertExpectations(t)

			require.Equal(t, test.expectedStatus, rec.Code)
			require.Equal(t, test.expectedCursor, rec.Header().Get("X-Cursor"))
		})
	}
}
