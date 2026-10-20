package usecase

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/situmorangbastian/skyros/orderservice/internal/models"
	"github.com/situmorangbastian/skyros/serviceutils/auth"
)

type fakeRepo struct {
	stored       models.Order
	lastFilter   models.Filter
	patchedOwner string
}

func (f *fakeRepo) Store(_ context.Context, o models.Order) (models.Order, error) {
	f.stored = o
	return o, nil
}
func (f *fakeRepo) Fetch(_ context.Context, filter models.Filter) ([]models.Order, error) {
	f.lastFilter = filter
	return nil, nil
}
func (f *fakeRepo) PatchStatus(_ context.Context, _, sellerID string, _ int) error {
	f.patchedOwner = sellerID
	return nil
}

type fakeProducts map[string]models.Product

func (f fakeProducts) FetchByIDs(context.Context, []string) (map[string]models.Product, error) {
	return f, nil
}

type fakeUsers struct{}

func (fakeUsers) FetchByIDs(context.Context, []string) (map[string]auth.Claims, error) {
	return map[string]auth.Claims{}, nil
}

// ctxWithClaims injects claims the same way AuthInterceptor does.
func ctxWithClaims(t *testing.T, c auth.Claims) context.Context {
	t.Helper()
	ctx := context.Background()
	// the claims key is private to package auth, so go through the interceptor
	return auth.WithClaimsForTest(ctx, c)
}

func newUsecase(repo *fakeRepo, products fakeProducts) OrderUsecase {
	return NewUsecase(repo, fakeUsers{}, products, zerolog.Nop())
}

func TestStore_ComputesTotalAndSetsBuyer(t *testing.T) {
	repo := &fakeRepo{}
	uc := newUsecase(repo, fakeProducts{
		"p1": {ID: "p1", Name: "A", Price: 100, Seller: auth.Claims{ID: "s1"}},
		"p2": {ID: "p2", Name: "B", Price: 50, Seller: auth.Claims{ID: "s1"}},
	})
	ctx := ctxWithClaims(t, auth.Claims{ID: "b1", Name: "Bob", Type: auth.UserBuyerType})

	got, err := uc.Store(ctx, models.Order{Items: []models.OrderProduct{
		{ProductID: "p1", Quantity: 2},
		{ProductID: "p2", Quantity: 1},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalPrice != 250 {
		t.Errorf("TotalPrice = %d, want 250", got.TotalPrice)
	}
	if got.Buyer.ID != "b1" || got.Buyer.Name != "Bob" || got.Seller.ID != "s1" {
		t.Errorf("unexpected parties: buyer=%+v seller=%+v", got.Buyer, got.Seller)
	}
}

func TestStore_RejectsMixedSellers(t *testing.T) {
	uc := newUsecase(&fakeRepo{}, fakeProducts{
		"p1": {ID: "p1", Name: "A", Price: 1, Seller: auth.Claims{ID: "s1"}},
		"p2": {ID: "p2", Name: "B", Price: 1, Seller: auth.Claims{ID: "s2"}},
	})
	ctx := ctxWithClaims(t, auth.Claims{ID: "b1", Type: auth.UserBuyerType})

	_, err := uc.Store(ctx, models.Order{Items: []models.OrderProduct{
		{ProductID: "p1", Quantity: 1}, {ProductID: "p2", Quantity: 1},
	}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", status.Code(err))
	}
}

func TestStore_OnlyBuyersCanOrder(t *testing.T) {
	uc := newUsecase(&fakeRepo{}, fakeProducts{})
	ctx := ctxWithClaims(t, auth.Claims{ID: "s1", Type: auth.UserSellerType})
	if _, err := uc.Store(ctx, models.Order{}); status.Code(err) != codes.NotFound {
		t.Fatalf("code = %v, want NotFound", status.Code(err))
	}
}

// Regression: buyers used to see every order because BuyerID was never applied.
func TestFetch_ScopesToCaller(t *testing.T) {
	repo := &fakeRepo{}
	uc := newUsecase(repo, fakeProducts{})

	if _, err := uc.Fetch(ctxWithClaims(t, auth.Claims{ID: "b1", Type: auth.UserBuyerType}), models.Filter{}); err != nil {
		t.Fatal(err)
	}
	if repo.lastFilter.BuyerID != "b1" || repo.lastFilter.SellerID != "" {
		t.Errorf("buyer filter = %+v", repo.lastFilter)
	}

	if _, err := uc.Fetch(ctxWithClaims(t, auth.Claims{ID: "s1", Type: auth.UserSellerType}), models.Filter{}); err != nil {
		t.Fatal(err)
	}
	if repo.lastFilter.SellerID != "s1" {
		t.Errorf("seller filter = %+v", repo.lastFilter)
	}
}

func TestPatchStatus_PassesCallerAsOwner(t *testing.T) {
	repo := &fakeRepo{}
	uc := newUsecase(repo, fakeProducts{})
	if err := uc.PatchStatus(ctxWithClaims(t, auth.Claims{ID: "s1", Type: auth.UserSellerType}), "o1", 1); err != nil {
		t.Fatal(err)
	}
	if repo.patchedOwner != "s1" {
		t.Errorf("owner = %q, want s1", repo.patchedOwner)
	}
}
