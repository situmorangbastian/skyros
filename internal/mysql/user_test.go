package mysql_test

import (
	"context"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/situmorangbastian/skyros"
	"github.com/situmorangbastian/skyros/internal/mysql"
	"github.com/situmorangbastian/skyros/testdata"
)

type userTestSuite struct {
	TestSuite
}

func (s *userTestSuite) seedUser(user skyros.User) {
	timeNow := time.Now()

	query, args, err := sq.Insert("user").
		Columns("id", "email", "name", "address", "password", "type", "created_time", "updated_time").
		Values(user.ID, user.Email, user.Name, user.Address, user.Password, user.Type, timeNow, timeNow).ToSql()
	require.NoError(s.T(), err)

	_, err = s.DBConn.ExecContext(context.TODO(), query, args...)
	require.NoError(s.T(), err)
}

func TestUserTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip user repository test")
	}

	suite.Run(t, new(userTestSuite))
}

func (s *userTestSuite) SetupTest() {
	_, err := s.DBConn.Exec("TRUNCATE user")
	require.NoError(s.T(), err)
}

func (s *userTestSuite) TestUser_Register() {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(s.T(), "user", &mockUser)

	s.T().Run("success", func(t *testing.T) {
		userRepo := mysql.NewUserRepository(s.DBConn)
		user, err := userRepo.Register(context.TODO(), mockUser)
		require.NoError(t, err)
		require.Equal(t, mockUser, user)
	})
}

func (s *userTestSuite) TestUser_GetUserByEmail() {
	var mockUser skyros.User
	testdata.GoldenJSONUnmarshal(s.T(), "user", &mockUser)

	s.seedUser(mockUser)

	s.T().Run("success", func(t *testing.T) {
		userRepo := mysql.NewUserRepository(s.DBConn)
		user, err := userRepo.GetUserByEmail(context.TODO(), mockUser.Email)
		require.NoError(t, err)
		require.Equal(t, mockUser, user)
	})

	s.T().Run("user not found", func(t *testing.T) {
		userRepo := mysql.NewUserRepository(s.DBConn)
		user, err := userRepo.GetUserByEmail(context.TODO(), "notfound@user.com")
		require.Error(t, err)
		require.Empty(t, user)
	})
}
