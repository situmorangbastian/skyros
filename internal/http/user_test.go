package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/situmorangbastian/skyros"
	"github.com/situmorangbastian/skyros/internal"
	handler "github.com/situmorangbastian/skyros/internal/http"
	"github.com/situmorangbastian/skyros/mocks"
	"github.com/situmorangbastian/skyros/testdata"
)

func TestUserHTTP_Login(t *testing.T) {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)

	tests := []struct {
		testName       string
		userService    testdata.FuncCall
		input          []byte
		expectedStatus int
	}{
		{
			testName: "success",
			userService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "email", "password"},
				Output: []interface{}{mockUser, nil},
			},
			input:          []byte(`{"email": "email", "password": "password"}`),
			expectedStatus: http.StatusOK,
		},
		{
			testName:       "error bad request input",
			input:          []byte(`{"email": "email", "password": "password",}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "error user service error not found",
			userService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "email", "password"},
				Output: []interface{}{skyros.User{}, skyros.ErrorNotFound("not found")},
			},
			input:          []byte(`{"email": "email", "password": "password"}`),
			expectedStatus: http.StatusNotFound,
		},
		{
			testName: "error user service unexpected error",
			userService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "email", "password"},
				Output: []interface{}{skyros.User{}, errors.New("unexpected error")},
			},
			input:          []byte(`{"email": "email", "password": "password"}`),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	e := echo.New()
	e.Use(handler.ErrorMiddleware())

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockUserService := new(mocks.UserService)
			if test.userService.Called {
				mockUserService.On("Login", test.userService.Input...).
					Return(test.userService.Output...).Once()
			}
			handler.NewUserHandler(e, mockUserService, "secret")

			req := httptest.NewRequest(echo.POST, "/login", strings.NewReader(string(test.input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			mockUserService.AssertExpectations(t)

			require.Equal(t, test.expectedStatus, rec.Code)

			if test.expectedStatus == http.StatusOK {
				var result map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &result)
				require.NoError(t, err)
				require.NotEmpty(t, result["access_token"])
			}
		})
	}
}

func TestUserHTTP_Register_Seller(t *testing.T) {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)

	mockUser.Type = "seller"

	tests := []struct {
		testName       string
		userService    testdata.FuncCall
		input          []byte
		expectedStatus int
	}{
		{
			testName: "success",
			userService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockUser},
				Output: []interface{}{mockUser, nil},
			},
			input:          testdata.GetGolden(t, "user"),
			expectedStatus: http.StatusCreated,
		},
		{
			testName:       "error bad request input",
			input:          []byte(`{"email": "email", "password": "password",}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "email already exist",
			userService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockUser},
				Output: []interface{}{skyros.User{}, skyros.ConflictError("email already exist")},
			},
			input:          testdata.GetGolden(t, "user"),
			expectedStatus: http.StatusConflict,
		},
		{
			testName: "error user service unexpected error",
			userService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockUser},
				Output: []interface{}{skyros.User{}, errors.New("unexpected error")},
			},
			input:          testdata.GetGolden(t, "user"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	e := echo.New()
	e.Use(handler.ErrorMiddleware())
	e.Validator = internal.NewValidator()

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockUserService := new(mocks.UserService)
			if test.userService.Called {
				mockUserService.On("Register", test.userService.Input...).
					Return(test.userService.Output...).Once()
			}
			handler.NewUserHandler(e, mockUserService, "secret")

			req := httptest.NewRequest(echo.POST, "/register/seller", strings.NewReader(string(test.input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			mockUserService.AssertExpectations(t)

			require.Equal(t, test.expectedStatus, rec.Code)

			if test.expectedStatus == http.StatusCreated {
				var result map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &result)
				require.NoError(t, err)
				require.NotEmpty(t, result["access_token"])
			}
		})
	}
}

func TestUserHTTP_Register_Buyer(t *testing.T) {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)

	mockUser.Type = "buyer"

	tests := []struct {
		testName       string
		userService    testdata.FuncCall
		input          []byte
		expectedStatus int
	}{
		{
			testName: "success",
			userService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockUser},
				Output: []interface{}{mockUser, nil},
			},
			input:          testdata.GetGolden(t, "user"),
			expectedStatus: http.StatusCreated,
		},
		{
			testName:       "error bad request input",
			input:          []byte(`{"email": "email", "password": "password",}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			testName: "email already exist",
			userService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockUser},
				Output: []interface{}{skyros.User{}, skyros.ConflictError("email already exist")},
			},
			input:          testdata.GetGolden(t, "user"),
			expectedStatus: http.StatusConflict,
		},
		{
			testName: "error user service unexpected error",
			userService: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockUser},
				Output: []interface{}{skyros.User{}, errors.New("unexpected error")},
			},
			input:          testdata.GetGolden(t, "user"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	e := echo.New()
	e.Use(handler.ErrorMiddleware())
	e.Validator = internal.NewValidator()

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockUserService := new(mocks.UserService)
			if test.userService.Called {
				mockUserService.On("Register", test.userService.Input...).
					Return(test.userService.Output...).Once()
			}
			handler.NewUserHandler(e, mockUserService, "secret")

			req := httptest.NewRequest(echo.POST, "/register/buyer", strings.NewReader(string(test.input)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			mockUserService.AssertExpectations(t)

			require.Equal(t, test.expectedStatus, rec.Code)

			if test.expectedStatus == http.StatusCreated {
				var result map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &result)
				require.NoError(t, err)
				require.NotEmpty(t, result["access_token"])
			}
		})
	}
}
