package product_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/situmorangbastian/skyros"
	"github.com/situmorangbastian/skyros/mocks"
	"github.com/situmorangbastian/skyros/product"
	"github.com/situmorangbastian/skyros/testdata"
)

func TestService_Store(t *testing.T) {
	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(t, "product", &mockProduct)

	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)

	mockUser.Type = "seller"
	mockProduct.Seller = mockUser

	tests := []struct {
		testName       string
		passedProduct  skyros.Product
		passedContext  context.Context
		repository     testdata.FuncCall
		expectedResult skyros.Product
		expectedError  error
	}{
		{
			testName:      "success",
			passedProduct: mockProduct,
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockProduct},
				Output: []interface{}{mockProduct, nil},
			},
			expectedResult: mockProduct,
			expectedError:  nil,
		},
		{
			testName:       "error parse custom context",
			passedProduct:  mockProduct,
			passedContext:  context.Background(),
			expectedResult: skyros.Product{},
			expectedError:  errors.New("invalid context"),
		},
		{
			testName:      "error from repository",
			passedProduct: mockProduct,
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockProduct},
				Output: []interface{}{skyros.Product{}, errors.New("unexpected error")},
			},
			expectedResult: skyros.Product{},
			expectedError:  errors.New("unexpected error"),
		},
	}

	repoMock := new(mocks.ProductRepository)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			if test.repository.Called {
				repoMock.On("Store", test.repository.Input...).
					Return(test.repository.Output...).Once()
			}

			service := product.NewService(repoMock)

			res, err := service.Store(test.passedContext, test.passedProduct)
			repoMock.AssertExpectations(t)

			if err != nil {
				require.EqualError(t, errors.Cause(err), test.expectedError.Error())
			}

			require.Equal(t, test.expectedResult, res)
		})
	}
}

func TestService_Get(t *testing.T) {
	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(t, "product", &mockProduct)

	tests := []struct {
		testName       string
		repository     testdata.FuncCall
		expectedResult skyros.Product
		expectedError  error
	}{
		{
			testName: "success",
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "product-id"},
				Output: []interface{}{mockProduct, nil},
			},
			expectedResult: mockProduct,
			expectedError:  nil,
		},
		{
			testName: "error from repository",
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "product-id"},
				Output: []interface{}{skyros.Product{}, errors.New("unexpected error")},
			},
			expectedResult: skyros.Product{},
			expectedError:  errors.New("unexpected error"),
		},
	}

	repoMock := new(mocks.ProductRepository)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			if test.repository.Called {
				repoMock.On("Get", test.repository.Input...).
					Return(test.repository.Output...).Once()
			}

			service := product.NewService(repoMock)

			res, err := service.Get(context.Background(), "product-id")
			repoMock.AssertExpectations(t)

			if err != nil {
				require.EqualError(t, errors.Cause(err), test.expectedError.Error())
			}

			require.Equal(t, test.expectedResult, res)
		})
	}
}

func TestService_Fetch(t *testing.T) {
	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(t, "product", &mockProduct)

	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)

	mockUser.Type = "buyer"
	mockProduct.Seller = mockUser

	passedFilterBuyer := skyros.Filter{
		Num:    1,
		Cursor: "next-cursor",
	}

	passedFilterSeller := passedFilterBuyer
	passedFilterSeller.SellerID = mockUser.ID

	mockUserSeller := mockUser
	mockUserSeller.Type = "seller"

	tests := []struct {
		testName       string
		passedFilter   skyros.Filter
		passedContext  context.Context
		repository     testdata.FuncCall
		expectedResult []skyros.Product
		expectedCursor string
		expectedError  error
	}{
		{
			testName:      "success with user buyer",
			passedFilter:  passedFilterBuyer,
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, passedFilterBuyer},
				Output: []interface{}{[]skyros.Product{mockProduct}, "next-cursor", nil},
			},
			expectedResult: []skyros.Product{mockProduct},
			expectedCursor: "next-cursor",
			expectedError:  nil,
		},
		{
			testName:      "success with user seller",
			passedFilter:  passedFilterSeller,
			passedContext: skyros.NewCustomContext(context.Background(), mockUserSeller),
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, passedFilterSeller},
				Output: []interface{}{[]skyros.Product{mockProduct}, "next-cursor", nil},
			},
			expectedResult: []skyros.Product{mockProduct},
			expectedCursor: "next-cursor",
			expectedError:  nil,
		},
		{
			testName:      "unexpected error from repository",
			passedFilter:  passedFilterBuyer,
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, passedFilterBuyer},
				Output: []interface{}{[]skyros.Product{}, "", errors.New("unexpected error")},
			},
			expectedResult: []skyros.Product{},
			expectedCursor: "",
			expectedError:  errors.New("unexpected error"),
		},
	}

	repoMock := new(mocks.ProductRepository)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			if test.repository.Called {
				repoMock.On("Fetch", test.repository.Input...).
					Return(test.repository.Output...).Once()
			}

			service := product.NewService(repoMock)

			res, nextCursor, err := service.Fetch(test.passedContext, test.passedFilter)
			repoMock.AssertExpectations(t)

			if err != nil {
				require.EqualError(t, errors.Cause(err), test.expectedError.Error())
			}

			require.Equal(t, test.expectedResult, res)
			require.Equal(t, test.expectedCursor, nextCursor)
		})
	}
}
