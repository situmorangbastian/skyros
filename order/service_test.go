package order_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/situmorangbastian/skyros"
	"github.com/situmorangbastian/skyros/mocks"
	"github.com/situmorangbastian/skyros/order"
	"github.com/situmorangbastian/skyros/testdata"
)

func TestService_Store(t *testing.T) {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)
	mockUser.Type = skyros.UserBuyerType
	mockUser.Password = ""

	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(t, "product", &mockProduct)

	var mockOrder skyros.Order
	testdata.GoldenJSONUnmarshal(t, "order", &mockOrder)
	mockOrder.Buyer = mockUser
	mockOrder.Items[0].Product = mockProduct

	mockUserSeller := mockUser
	mockUserSeller.Type = skyros.UserSellerType

	tests := []struct {
		testName        string
		passeOrder      skyros.Order
		passedContext   context.Context
		productService  []testdata.FuncCall
		orderRepository testdata.FuncCall
		expectedResult  skyros.Order
		expectedError   error
	}{
		{
			testName:      "success",
			passeOrder:    mockOrder,
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			productService: []testdata.FuncCall{
				{
					Called: true,
					Input:  []interface{}{mock.Anything, mockProduct.ID},
					Output: []interface{}{mockProduct, nil},
				},
			},
			orderRepository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockOrder},
				Output: []interface{}{mockOrder, nil},
			},
			expectedResult: mockOrder,
			expectedError:  nil,
		},
		{
			testName:       "error parse custom context",
			passedContext:  context.Background(),
			expectedResult: skyros.Order{},
			expectedError:  errors.New("invalid context"),
		},
		{
			testName:       "error, user not buyer",
			passeOrder:     mockOrder,
			passedContext:  skyros.NewCustomContext(context.Background(), mockUserSeller),
			expectedResult: skyros.Order{},
			expectedError:  skyros.ErrorNotFound("not found"),
		},
		{
			testName:      "error from repository",
			passeOrder:    mockOrder,
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			productService: []testdata.FuncCall{
				{
					Called: true,
					Input:  []interface{}{mock.Anything, mockProduct.ID},
					Output: []interface{}{mockProduct, nil},
				},
			},
			orderRepository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockOrder},
				Output: []interface{}{skyros.Order{}, errors.New("unexpected error")},
			},
			expectedResult: skyros.Order{},
			expectedError:  errors.New("unexpected error"),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			orderRepoMock := new(mocks.OrderRepository)
			productServiceMock := new(mocks.ProductService)

			if test.orderRepository.Called {
				orderRepoMock.On("Store", test.orderRepository.Input...).
					Return(test.orderRepository.Output...).Once()
			}

			for _, productService := range test.productService {
				if productService.Called {
					productServiceMock.On("Get", productService.Input...).
						Return(productService.Output...).Once()
				}
			}

			service := order.NewService(orderRepoMock, productServiceMock)

			res, err := service.Store(test.passedContext, test.passeOrder)
			orderRepoMock.AssertExpectations(t)
			productServiceMock.AssertExpectations(t)

			if err != nil {
				require.EqualError(t, errors.Cause(err), test.expectedError.Error())
			}

			require.Equal(t, test.expectedResult, res)
		})
	}
}

func TestService_Get(t *testing.T) {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)
	mockUser.Type = skyros.UserBuyerType
	mockUser.Password = ""

	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(t, "product", &mockProduct)

	var mockOrder skyros.Order
	testdata.GoldenJSONUnmarshal(t, "order", &mockOrder)
	mockOrder.Buyer = mockUser
	mockOrder.Items[0].Product = mockProduct

	mockUserSeller := mockUser
	mockUserSeller.Type = skyros.UserSellerType

	unsupportedUser := mockUser
	unsupportedUser.Type = ""

	tests := []struct {
		testName        string
		passedContext   context.Context
		orderRepository testdata.FuncCall
		productService  []testdata.FuncCall
		expectedResult  skyros.Order
		expectedError   error
	}{
		{
			testName:      "success with user buyer",
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			orderRepository: testdata.FuncCall{
				Called: true,
				Input: []interface{}{mock.Anything, skyros.Filter{
					BuyerID: mockUser.ID,
					OrderID: mockOrder.ID,
				}},
				Output: []interface{}{[]skyros.Order{mockOrder}, "", nil},
			},
			productService: []testdata.FuncCall{
				{
					Called: true,
					Input:  []interface{}{mock.Anything, mockProduct.ID},
					Output: []interface{}{mockProduct, nil},
				},
			},
			expectedResult: mockOrder,
			expectedError:  nil,
		},
		{
			testName:       "error parse custom context",
			passedContext:  context.Background(),
			expectedResult: skyros.Order{},
			expectedError:  errors.New("invalid context"),
		},
		{
			testName:      "success with user seller",
			passedContext: skyros.NewCustomContext(context.Background(), mockUserSeller),
			orderRepository: testdata.FuncCall{
				Called: true,
				Input: []interface{}{mock.Anything, skyros.Filter{
					SellerID: mockUserSeller.ID,
					OrderID:  mockOrder.ID,
				}},
				Output: []interface{}{[]skyros.Order{mockOrder}, "", nil},
			},
			productService: []testdata.FuncCall{
				{
					Called: true,
					Input:  []interface{}{mock.Anything, mockProduct.ID},
					Output: []interface{}{mockProduct, nil},
				},
			},
			expectedResult: mockOrder,
			expectedError:  nil,
		},
		{
			testName:       "error with unsupported type user",
			passedContext:  skyros.NewCustomContext(context.Background(), unsupportedUser),
			expectedResult: skyros.Order{},
			expectedError:  skyros.ErrorNotFound("not found"),
		},
		{
			testName:      "error from repository",
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			orderRepository: testdata.FuncCall{
				Called: true,
				Input: []interface{}{mock.Anything, skyros.Filter{
					BuyerID: mockUser.ID,
					OrderID: mockOrder.ID,
				}},
				Output: []interface{}{[]skyros.Order{}, "", errors.New("unexpected error")},
			},
			expectedResult: skyros.Order{},
			expectedError:  errors.New("unexpected error"),
		},
		{
			testName:      "error order not found",
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			orderRepository: testdata.FuncCall{
				Called: true,
				Input: []interface{}{mock.Anything, skyros.Filter{
					BuyerID: mockUser.ID,
					OrderID: mockOrder.ID,
				}},
				Output: []interface{}{[]skyros.Order{}, "", nil},
			},
			expectedResult: skyros.Order{},
			expectedError:  skyros.ErrorNotFound("not found"),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			orderRepoMock := new(mocks.OrderRepository)
			productServiceMock := new(mocks.ProductService)

			if test.orderRepository.Called {
				orderRepoMock.On("Fetch", test.orderRepository.Input...).
					Return(test.orderRepository.Output...).Once()
			}

			for _, productService := range test.productService {
				if productService.Called {
					productServiceMock.On("Get", productService.Input...).
						Return(productService.Output...).Once()
				}
			}

			service := order.NewService(orderRepoMock, productServiceMock)

			res, err := service.Get(test.passedContext, mockOrder.ID)
			orderRepoMock.AssertExpectations(t)
			productServiceMock.AssertExpectations(t)

			if err != nil {
				require.EqualError(t, errors.Cause(err), test.expectedError.Error())
			}

			require.Equal(t, test.expectedResult, res)
		})
	}
}

func TestService_Fetch(t *testing.T) {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)
	mockUser.Type = skyros.UserBuyerType
	mockUser.Password = ""

	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(t, "product", &mockProduct)

	var mockOrder skyros.Order
	testdata.GoldenJSONUnmarshal(t, "order", &mockOrder)
	mockOrder.Buyer = mockUser
	mockOrder.Items[0].Product = mockProduct

	mockUserSeller := mockUser
	mockUserSeller.Type = skyros.UserSellerType

	unsupportedUser := mockUser
	unsupportedUser.Type = ""

	filter := skyros.Filter{
		Num: 20,
	}

	filterForUserBuyer := filter
	filterForUserBuyer.BuyerID = mockUser.ID

	filterForUserSeller := filter
	filterForUserSeller.BuyerID = mockUserSeller.ID

	tests := []struct {
		testName           string
		passedContext      context.Context
		passedFilter       skyros.Filter
		orderRepository    testdata.FuncCall
		productService     []testdata.FuncCall
		expectedResult     []skyros.Order
		expectedNextCursor string
		expectedError      error
	}{
		{
			testName:      "success with user buyer",
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			orderRepository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, filterForUserBuyer},
				Output: []interface{}{[]skyros.Order{mockOrder}, "next-cursor", nil},
			},
			productService: []testdata.FuncCall{
				{
					Called: true,
					Input:  []interface{}{mock.Anything, mockProduct.ID},
					Output: []interface{}{mockProduct, nil},
				},
			},
			passedFilter:       filterForUserBuyer,
			expectedResult:     []skyros.Order{mockOrder},
			expectedNextCursor: "next-cursor",
			expectedError:      nil,
		},
		{
			testName:      "success with user seller",
			passedContext: skyros.NewCustomContext(context.Background(), mockUserSeller),
			orderRepository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, filterForUserSeller},
				Output: []interface{}{[]skyros.Order{mockOrder}, "next-cursor", nil},
			},
			productService: []testdata.FuncCall{
				{
					Called: true,
					Input:  []interface{}{mock.Anything, mockProduct.ID},
					Output: []interface{}{mockProduct, nil},
				},
			},
			passedFilter:       filterForUserSeller,
			expectedResult:     []skyros.Order{mockOrder},
			expectedNextCursor: "next-cursor",
			expectedError:      nil,
		},
		{
			testName:       "error parse custom context",
			passedContext:  context.Background(),
			expectedResult: []skyros.Order{},
			expectedError:  errors.New("invalid context"),
		},
		{
			testName:       "error with unsupported type user",
			passedContext:  skyros.NewCustomContext(context.Background(), unsupportedUser),
			expectedResult: []skyros.Order{},
			expectedError:  skyros.ErrorNotFound("not found"),
		},
		{
			testName:      "error from repository",
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			orderRepository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, filterForUserBuyer},
				Output: []interface{}{[]skyros.Order{}, "", errors.New("unexpected error")},
			},
			passedFilter:   filterForUserBuyer,
			expectedResult: []skyros.Order{},
			expectedError:  errors.New("unexpected error"),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			orderRepoMock := new(mocks.OrderRepository)
			productServiceMock := new(mocks.ProductService)

			if test.orderRepository.Called {
				orderRepoMock.On("Fetch", test.orderRepository.Input...).
					Return(test.orderRepository.Output...).Once()
			}

			for _, productService := range test.productService {
				if productService.Called {
					productServiceMock.On("Get", productService.Input...).
						Return(productService.Output...).Once()
				}
			}

			service := order.NewService(orderRepoMock, productServiceMock)

			res, cursor, err := service.Fetch(test.passedContext, test.passedFilter)
			orderRepoMock.AssertExpectations(t)
			productServiceMock.AssertExpectations(t)

			if err != nil {
				require.EqualError(t, errors.Cause(err), test.expectedError.Error())
			}

			require.Equal(t, test.expectedResult, res)
			require.Equal(t, test.expectedNextCursor, cursor)
		})
	}
}

func TestService_PatchStatus(t *testing.T) {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)
	mockUser.Type = skyros.UserSellerType
	mockUser.Password = ""

	mockUserBuyer := mockUser
	mockUserBuyer.Type = skyros.UserBuyerType

	tests := []struct {
		testName       string
		passedContext  context.Context
		repository     testdata.FuncCall
		expectedResult skyros.Order
		expectedError  error
	}{
		{
			testName:      "success with user seller",
			passedContext: skyros.NewCustomContext(context.Background(), mockUser),
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "order-id", 1},
				Output: []interface{}{nil},
			},
			expectedError: nil,
		},
		{
			testName:       "error, user not seller",
			passedContext:  skyros.NewCustomContext(context.Background(), mockUserBuyer),
			expectedResult: skyros.Order{},
			expectedError:  skyros.ErrorNotFound("not found"),
		},
		{
			testName:       "error parse custom context",
			passedContext:  context.Background(),
			expectedResult: skyros.Order{},
			expectedError:  errors.New("invalid context"),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			repoMock := new(mocks.OrderRepository)

			if test.repository.Called {
				repoMock.On("PatchStatus", test.repository.Input...).
					Return(test.repository.Output...).Once()
			}

			service := order.NewService(repoMock, nil)

			err := service.PatchStatus(test.passedContext, "order-id", 1)
			repoMock.AssertExpectations(t)

			if err != nil {
				require.EqualError(t, errors.Cause(err), test.expectedError.Error())
			}

		})
	}
}
