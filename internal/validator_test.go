package internal_test

import (
	"testing"

	"github.com/situmorangbastian/skyros"
	"github.com/situmorangbastian/skyros/internal"
	"github.com/situmorangbastian/skyros/testdata"
	"github.com/stretchr/testify/require"
)

func TestValidate_User(t *testing.T) {
	t.Run("success with no error", func(t *testing.T) {
		var user skyros.User
		testdata.GoldenJSONUnmarshal(t, "user", &user)

		v := internal.NewValidator()
		err := v.Validate(user)
		require.Nil(t, err)
	})

	t.Run("error name is required", func(t *testing.T) {
		var user skyros.User
		testdata.GoldenJSONUnmarshal(t, "user", &user)
		user.Name = ""

		v := internal.NewValidator()
		err := v.Validate(user)
		require.Error(t, err)
		require.Equal(t, "name required", err.Error())
	})

	t.Run("error email is required", func(t *testing.T) {
		var user skyros.User
		testdata.GoldenJSONUnmarshal(t, "user", &user)
		user.Email = ""

		v := internal.NewValidator()
		err := v.Validate(user)
		require.Error(t, err)
		require.Equal(t, "email required", err.Error())
	})

	t.Run("error email is invalid", func(t *testing.T) {
		var user skyros.User
		testdata.GoldenJSONUnmarshal(t, "user", &user)
		user.Email = "user"

		v := internal.NewValidator()
		err := v.Validate(user)
		require.Error(t, err)
		require.Equal(t, "invalid email", err.Error())
	})

	t.Run("password is required", func(t *testing.T) {
		var user skyros.User
		testdata.GoldenJSONUnmarshal(t, "user", &user)
		user.Password = ""

		v := internal.NewValidator()
		err := v.Validate(user)
		require.Error(t, err)
		require.Equal(t, "password required", err.Error())
	})

	t.Run("address is required", func(t *testing.T) {
		var user skyros.User
		testdata.GoldenJSONUnmarshal(t, "user", &user)
		user.Address = ""

		v := internal.NewValidator()
		err := v.Validate(user)
		require.Error(t, err)
		require.Equal(t, "address required", err.Error())
	})
}
