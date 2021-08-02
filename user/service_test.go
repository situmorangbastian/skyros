package user_test

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/situmorangbastian/skyros"
	"github.com/situmorangbastian/skyros/mocks"
	"github.com/situmorangbastian/skyros/testdata"
	"github.com/situmorangbastian/skyros/user"
)

func TestService_Login(t *testing.T) {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)

	tests := []struct {
		testName       string
		passedEmail    string
		passedPassword string
		repository     testdata.FuncCall
		expectedResult skyros.User
		expectedError  error
	}{
		{
			testName:       "success",
			passedEmail:    "user@example.com",
			passedPassword: "password123#",
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "user@example.com"},
				Output: []interface{}{mockUser, nil},
			},
			expectedResult: mockUser,
			expectedError:  nil,
		},
		{
			testName:       "error invalid password",
			passedEmail:    "user@example.com",
			passedPassword: "password",
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "user@example.com"},
				Output: []interface{}{mockUser, nil},
			},
			expectedError: skyros.ErrorNotFound("user not found"),
		},
		{
			testName:       "with unexpected error from user repository",
			passedEmail:    "user@example.com",
			passedPassword: "password123#",
			repository: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, "user@example.com"},
				Output: []interface{}{mockUser, errors.New("unexpected error")},
			},
			expectedError: errors.New("unexpected error"),
		},
	}

	repoMock := new(mocks.UserRepository)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			if test.repository.Called {
				repoMock.On("GetUser", test.repository.Input...).
					Return(test.repository.Output...).Once()
			}

			service := user.NewService(repoMock)

			res, err := service.Login(context.Background(), test.passedEmail, test.passedPassword)
			repoMock.AssertExpectations(t)

			if err != nil {
				require.EqualError(t, errors.Cause(err), test.expectedError.Error())
				return
			}

			require.Equal(t, test.expectedResult, res)
		})
	}
}

func TestService_Register(t *testing.T) {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(t, "user", &mockUser)

	passedUser := mockUser
	passedUser.Password = "password123#"

	tests := []struct {
		testName       string
		repoGetUser    testdata.FuncCall
		repoRegister   testdata.FuncCall
		expectedResult skyros.User
		expectedError  error
	}{
		{
			testName: "success",
			repoGetUser: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockUser.Email},
				Output: []interface{}{skyros.User{}, skyros.ErrorNotFound("user not found")},
			},
			repoRegister: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mock.AnythingOfType("skyros.User")},
				Output: []interface{}{mockUser, nil},
			},
			expectedResult: mockUser,
			expectedError:  nil,
		},
		{
			testName: "error email already exist",
			repoGetUser: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockUser.Email},
				Output: []interface{}{mockUser, nil},
			},
			expectedError: skyros.ConflictError("email already exist"),
		},
		{
			testName: "error get user by email",
			repoGetUser: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockUser.Email},
				Output: []interface{}{mockUser, errors.New("unexpected error")},
			},
			expectedError: errors.New("unexpected error"),
		},
		{
			testName: "unexpected error",
			repoGetUser: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mockUser.Email},
				Output: []interface{}{skyros.User{}, skyros.ErrorNotFound("user not found")},
			},
			repoRegister: testdata.FuncCall{
				Called: true,
				Input:  []interface{}{mock.Anything, mock.AnythingOfType("skyros.User")},
				Output: []interface{}{skyros.User{}, errors.New("unexpected error")},
			},
			expectedError: errors.New("unexpected error"),
		},
	}

	repoMock := new(mocks.UserRepository)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			if test.repoGetUser.Called {
				repoMock.On("GetUser", test.repoGetUser.Input...).
					Return(test.repoGetUser.Output...).Once()
			}

			if test.repoRegister.Called {
				repoMock.On("Register", test.repoRegister.Input...).
					Return(test.repoRegister.Output...).Once()
			}

			service := user.NewService(repoMock)

			res, err := service.Register(context.Background(), passedUser)
			repoMock.AssertExpectations(t)

			if err != nil {
				require.EqualError(t, errors.Cause(err), test.expectedError.Error())
				return
			}

			require.Equal(t, test.expectedResult, res)
		})
	}
}
