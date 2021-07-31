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

type productTestSuite struct {
	TestSuite
}

func (s *productTestSuite) seedProduct(product skyros.Product) {
	query, args, err := sq.Insert("product").
		Columns("id", "name", "description", "price", "seller_id", "created_time", "updated_time").
		Values(product.ID, product.Name, product.Description, product.Price, product.Seller.ID, product.CreatedTime, product.UpdatedTime).ToSql()
	require.NoError(s.T(), err)

	_, err = s.DBConn.ExecContext(context.TODO(), query, args...)
	require.NoError(s.T(), err)
}

func TestProductTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip user repository test")
	}

	suite.Run(t, new(productTestSuite))
}

func (s *productTestSuite) SetupTest() {
	_, err := s.DBConn.Exec("TRUNCATE product")
	require.NoError(s.T(), err)
}

func (s *productTestSuite) TestProduct_Store() {
	timeNow, err := time.Parse("2006-01-02T:15:04:05+07:00", "2018-06-25T:10:00:00+07:00")
	require.NoError(s.T(), err)

	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(s.T(), "product", &mockProduct)

	mockProduct.Seller.ID = "seller-id"
	mockProduct.CreatedTime = timeNow
	mockProduct.UpdatedTime = timeNow

	productRepo := mysql.NewProductRepository(s.DBConn)
	product, err := productRepo.Store(context.TODO(), mockProduct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), mockProduct.Name, product.Name)
	require.Equal(s.T(), mockProduct.Description, product.Description)
	require.Equal(s.T(), mockProduct.Price, product.Price)
	require.Equal(s.T(), mockProduct.Seller.ID, product.Seller.ID)
}

func (s *productTestSuite) TestProduct_Get() {
	timeNow, err := time.Parse("2006-01-02T:15:04:05+07:00", "2018-06-25T:10:00:00+07:00")
	require.NoError(s.T(), err)

	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(s.T(), "product", &mockProduct)

	mockProduct.Seller.ID = "seller-id"
	mockProduct.CreatedTime = timeNow
	mockProduct.UpdatedTime = timeNow

	s.seedProduct(mockProduct)

	s.T().Run("success", func(t *testing.T) {
		productRepo := mysql.NewProductRepository(s.DBConn)
		product, err := productRepo.Get(context.TODO(), mockProduct.ID)
		require.NoError(s.T(), err)
		require.Equal(s.T(), mockProduct, product)
	})

	s.T().Run("error not found", func(t *testing.T) {
		productRepo := mysql.NewProductRepository(s.DBConn)
		product, err := productRepo.Get(context.TODO(), "other-product-id")
		require.EqualError(s.T(), err, "product not found")
		require.Empty(s.T(), product)
	})
}

func (s *productTestSuite) TestProduct_Fetch() {
	timeNow, err := time.Parse("2006-01-02T:15:04:05+07:00", "2018-06-25T:10:00:00+07:00")
	require.NoError(s.T(), err)

	var mockProduct skyros.Product
	testdata.GoldenJSONUnmarshal(s.T(), "product", &mockProduct)

	mockProduct.Seller.ID = "seller-id"
	mockProduct.CreatedTime = timeNow
	mockProduct.UpdatedTime = timeNow

	s.seedProduct(mockProduct)

	otherMockProduct := mockProduct
	otherMockProduct.ID = "other-product-id"
	otherMockProduct.Name = "Gas"
	otherMockProduct.Description = "Gas"
	otherMockProduct.CreatedTime = timeNow.Add(2 * time.Second)
	otherMockProduct.UpdatedTime = timeNow.Add(2 * time.Second)

	s.seedProduct(otherMockProduct)

	s.T().Run("success with filter num", func(t *testing.T) {
		productRepo := mysql.NewProductRepository(s.DBConn)
		products, nextCursor, err := productRepo.Fetch(context.TODO(), skyros.Filter{
			Num: 1,
		})
		require.NoError(s.T(), err)
		require.Equal(s.T(), "MjAxOC0wNi0yNVQxMDowMDowMlo=", nextCursor)
		require.Equal(s.T(), []skyros.Product{otherMockProduct}, products)
	})

	s.T().Run("success with filter num and cursor", func(t *testing.T) {
		productRepo := mysql.NewProductRepository(s.DBConn)
		products, nextCursor, err := productRepo.Fetch(context.TODO(), skyros.Filter{
			Num:    1,
			Cursor: "MjAxOC0wNi0yNVQxMDowMDowMlo=",
		})
		require.NoError(s.T(), err)
		require.Equal(s.T(), "MjAxOC0wNi0yNVQxMDowMDowMFo=", nextCursor)
		require.Equal(s.T(), []skyros.Product{mockProduct}, products)
	})

	s.T().Run("success with filter num and search", func(t *testing.T) {
		productRepo := mysql.NewProductRepository(s.DBConn)
		products, nextCursor, err := productRepo.Fetch(context.TODO(), skyros.Filter{
			Num:    1,
			Search: "minyak",
		})
		require.NoError(s.T(), err)
		require.Equal(s.T(), "MjAxOC0wNi0yNVQxMDowMDowMFo=", nextCursor)
		require.Equal(s.T(), []skyros.Product{mockProduct}, products)
	})

	s.T().Run("success with filter num, search, and seller id", func(t *testing.T) {
		productRepo := mysql.NewProductRepository(s.DBConn)
		products, nextCursor, err := productRepo.Fetch(context.TODO(), skyros.Filter{
			Num:      1,
			Search:   "minyak",
			SellerID: "seller-id",
		})
		require.NoError(s.T(), err)
		require.Equal(s.T(), "MjAxOC0wNi0yNVQxMDowMDowMFo=", nextCursor)
		require.Equal(s.T(), []skyros.Product{mockProduct}, products)
	})

	s.T().Run("error invalid cursor", func(t *testing.T) {
		productRepo := mysql.NewProductRepository(s.DBConn)
		products, nextCursor, err := productRepo.Fetch(context.TODO(), skyros.Filter{
			Num:    1,
			Cursor: "invalid-cursor",
		})
		require.Error(s.T(), err)
		require.Empty(s.T(), nextCursor)
		require.Empty(s.T(), products)
	})
}
