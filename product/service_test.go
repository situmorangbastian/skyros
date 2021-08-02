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

	mockUser.Type = skyros.UserSellerType
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

			service := product.NewService(repoMock, nil)

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

	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)

	mockProduct.Seller = mockUser

	tests := []struct {
		testName       string
		productRepo    testdata.FuncCall
		userRepo       testdata.FuncCall
		expectedResult skyros.Product
		expectedError  error
	}{
		{
			testName: "success",
			productRepo: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "product-id"},
				Output: []interface{}{mockProduct, nil},
			},
			userRepo: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockProduct.Seller.ID},
				Output: []interface{}{mockUser, nil},
			},
			expectedResult: mockProduct,
			expectedError:  nil,
		},
		{
			testName: "error from repository",
			productRepo: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "product-id"},
				Output: []interface{}{skyros.Product{}, errors.New("unexpected error")},
			},
			expectedResult: skyros.Product{},
			expectedError:  errors.New("unexpected error"),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			productRepoMock := new(mocks.ProductRepository)
			userRepoMock := new(mocks.UserRepository)

			if test.productRepo.Called {
				productRepoMock.On("Get", test.productRepo.Input...).
					Return(test.productRepo.Output...).Once()
			}

			if test.userRepo.Called {
				userRepoMock.On("GetUser", test.userRepo.Input...).
					Return(test.userRepo.Output...).Once()
			}

			service := product.NewService(productRepoMock, userRepoMock)

			res, err := service.Get(context.Background(), "product-id")
			productRepoMock.AssertExpectations(t)
			userRepoMock.AssertExpectations(t)

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

	mockUser.Type = skyros.UserBuyerType
	mockProduct.Seller = mockUser

	passedFilterBuyer := skyros.Filter{
		Num:    1,
		Cursor: "next-cursor",
	}

	passedFilterSeller := passedFilterBuyer
	passedFilterSeller.SellerID = mockUser.ID

	mockUserSeller := mockUser
	mockUserSeller.Type = skyros.UserSellerType

	tests := []struct {
		testName          string
		passedFilter      skyros.Filter
		passedContext     context.Context
		productRepository testdata.FuncCall
		userRepo          []testdata.FuncCall
		expectedResult    []skyros.Product
		expectedCursor    string
		expectedError     error
	}{
		{
			testName:      "success with user buyer",
			passedFilter:  passedFilterBuyer,
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			productRepository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, passedFilterBuyer},
				Output: []interface{}{[]skyros.Product{mockProduct}, "next-cursor", nil},
			},
			userRepo: []testdata.FuncCall{
				{
					Called: true,
					Input:  []interface{}{mock.Anything, mockProduct.Seller.ID},
					Output: []interface{}{mockUser, nil},
				},
			},
			expectedResult: []skyros.Product{mockProduct},
			expectedCursor: "next-cursor",
			expectedError:  nil,
		},
		{
			testName:      "success with user seller",
			passedFilter:  passedFilterSeller,
			passedContext: skyros.NewCustomContext(context.Background(), mockUserSeller),
			productRepository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, passedFilterSeller},
				Output: []interface{}{[]skyros.Product{mockProduct}, "next-cursor", nil},
			},
			userRepo: []testdata.FuncCall{
				{
					Called: true,
					Input:  []interface{}{mock.Anything, mockProduct.Seller.ID},
					Output: []interface{}{mockUser, nil},
				},
			},
			expectedResult: []skyros.Product{mockProduct},
			expectedCursor: "next-cursor",
			expectedError:  nil,
		},
		{
			testName:      "unexpected error from repository",
			passedFilter:  passedFilterBuyer,
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			productRepository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, passedFilterBuyer},
				Output: []interface{}{[]skyros.Product{}, "", errors.New("unexpected error")},
			},
			expectedResult: []skyros.Product{},
			expectedCursor: "",
			expectedError:  errors.New("unexpected error"),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			productRepoMock := new(mocks.ProductRepository)
			userRepoMock := new(mocks.UserRepository)

			if test.productRepository.Called {
				productRepoMock.On("Fetch", test.productRepository.Input...).
					Return(test.productRepository.Output...).Once()
			}

			for _, userRepo := range test.userRepo {
				if userRepo.Called {
					userRepoMock.On("GetUser", userRepo.Input...).
						Return(userRepo.Output...).Once()
				}
			}

			service := product.NewService(productRepoMock, userRepoMock)

			res, nextCursor, err := service.Fetch(test.passedContext, test.passedFilter)
			productRepoMock.AssertExpectations(t)
			userRepoMock.AssertExpectations(t)

			if err != nil {
				require.EqualError(t, errors.Cause(err), test.expectedError.Error())
			}

			require.Equal(t, test.expectedResult, res)
			require.Equal(t, test.expectedCursor, nextCursor)
		})
	}
}
