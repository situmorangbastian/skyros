package usecase

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/situmorangbastian/skyros/userservice/internal/models"
	"github.com/situmorangbastian/skyros/userservice/internal/repository"
)

type fakeRepo struct{ users map[string]models.User }

func (f fakeRepo) Register(_ context.Context, u models.User) (models.User, error) { return u, nil }
func (f fakeRepo) GetUserByEmail(_ context.Context, email string) (models.User, error) {
	u, ok := f.users[email]
	if !ok {
		return models.User{}, repository.ErrNotFound
	}
	return u, nil
}
func (f fakeRepo) FetchUsersByIDs(context.Context, []string) (map[string]models.User, error) {
	return nil, nil
}

// Unknown email and wrong password must be indistinguishable to the caller.
func TestLogin_InvalidCredentialsAreUnauthenticated(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	uc := NewUserUsecase(fakeRepo{users: map[string]models.User{
		"a@x.com": {ID: "1", Email: "a@x.com", Password: string(hash)},
	}}, zerolog.Nop())

	for name, tc := range map[string]struct{ email, password string }{
		"unknown email":  {"nobody@x.com", "right"},
		"wrong password": {"a@x.com", "wrong"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := uc.Login(context.Background(), tc.email, tc.password)
			st, _ := status.FromError(err)
			if st.Code() != codes.Unauthenticated || st.Message() != "invalid email or password" {
				t.Fatalf("got %v, want Unauthenticated invalid email or password", err)
			}
		})
	}

	if _, err := uc.Login(context.Background(), "a@x.com", "right"); err != nil {
		t.Fatalf("valid login failed: %v", err)
	}
}
