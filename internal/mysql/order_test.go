package mysql_test

import (
	"context"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/situmorangbastian/skyros"
	"github.com/situmorangbastian/skyros/internal/mysql"
	"github.com/situmorangbastian/skyros/testdata"
)

type orderTestSuite struct {
	TestSuite
}

func (s *orderTestSuite) seedOrder(order skyros.Order) {
	ctx := context.TODO()

	tx, err := s.TestSuite.DBConn.BeginTx(ctx, nil)
	require.NoError(s.T(), err)

	query, args, err := sq.Insert("orders").
		Columns("id", "buyer_id", "seller_id", "description", "source_address", "destination_address", "total_price", "status", "created_time", "updated_time").
		Values(order.ID, order.Buyer.ID, order.Seller.ID, order.Description, order.SourceAddress, order.DestinationAddress, order.TotalPrice, order.Status, order.CreatedTime, order.UpdatedTime).ToSql()
	require.NoError(s.T(), err)

	stmt, err := tx.PrepareContext(ctx, query)
	require.NoError(s.T(), err)
	defer func() {
		err := stmt.Close()
		require.NoError(s.T(), err)
	}()

	_, err = stmt.ExecContext(ctx, args...)
	require.NoError(s.T(), err)

	for _, orderItem := range order.Items {
		query, args, err := sq.Insert("orders_product").
			Columns("id", "order_id", "product_id", "quantity", "created_time", "updated_time").
			Values(uuid.New().String(), order.ID, orderItem.ProductID, orderItem.Quantity, order.CreatedTime, order.UpdatedTime).ToSql()
		require.NoError(s.T(), err)

		stmt, err := tx.PrepareContext(ctx, query)
		require.NoError(s.T(), err)
		defer func() {
			err := stmt.Close()
			require.NoError(s.T(), err)
		}()

		_, err = stmt.ExecContext(ctx, args...)
		require.NoError(s.T(), err)
	}

	err = tx.Commit()
	require.NoError(s.T(), err)
}

func TestOrderTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip user repository test")
	}

	suite.Run(t, new(orderTestSuite))
}

func (s *orderTestSuite) SetupTest() {
	_, err := s.DBConn.Exec("TRUNCATE product")
	require.NoError(s.T(), err)
}

func (s *orderTestSuite) TestOrder_Store() {
	var mockOrder skyros.Order
	testdata.GoldenJSONUnmarshal(s.T(), "order", &mockOrder)

	mockOrder.Seller.ID = "seller-id"
	mockOrder.Buyer.ID = "buyer-id"

	orderRepo := mysql.NewOrderRepository(s.DBConn)
	order, err := orderRepo.Store(context.TODO(), mockOrder)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), order.ID)
	require.Equal(s.T(), len(mockOrder.Items), len(mockOrder.Items))
}

func (s *orderTestSuite) TestOrder_Fetch() {
	timeNow, err := time.Parse("2006-01-02T:15:04:05+07:00", "2018-06-25T:10:00:00+07:00")
	require.NoError(s.T(), err)

	var mockOrder skyros.Order
	testdata.GoldenJSONUnmarshal(s.T(), "order", &mockOrder)

	mockOrder.Seller.ID = "seller-id"
	mockOrder.Buyer.ID = "buyer-id"
	mockOrder.CreatedTime = timeNow
	mockOrder.UpdatedTime = timeNow

	s.seedOrder(mockOrder)

	otherMockOrder := mockOrder
	otherMockOrder.ID = "order-id-2"
	otherMockOrder.CreatedTime = timeNow.Add(2 * time.Second)
	otherMockOrder.UpdatedTime = timeNow.Add(2 * time.Second)

	s.seedOrder(otherMockOrder)

	s.T().Run("success with filter num", func(t *testing.T) {
		orderRepo := mysql.NewOrderRepository(s.DBConn)
		orders, cursor, err := orderRepo.Fetch(context.TODO(), skyros.Filter{
			Num: 1,
		})
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), orders)
		require.NotEmpty(s.T(), orders[0].Items)
		require.Equal(s.T(), mockOrder.Items[0].ProductID, orders[0].Items[0].ProductID)
		require.Equal(s.T(), mockOrder.Items[0].Quantity, orders[0].Items[0].Quantity)
		require.Equal(s.T(), "MjAxOC0wNi0yNVQxMDowMDowMlo=", cursor)
	})

	s.T().Run("success with filter num and cursor", func(t *testing.T) {
		orderRepo := mysql.NewOrderRepository(s.DBConn)
		orders, nextCursor, err := orderRepo.Fetch(context.TODO(), skyros.Filter{
			Num:    1,
			Cursor: "MjAxOC0wNi0yNVQxMDowMDowMlo=",
		})
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), orders)
		require.NotEmpty(s.T(), orders[0].Items)
		require.Equal(s.T(), mockOrder.Items[0].ProductID, orders[0].Items[0].ProductID)
		require.Equal(s.T(), mockOrder.Items[0].Quantity, orders[0].Items[0].Quantity)
		require.Equal(s.T(), "MjAxOC0wNi0yNVQxMDowMDowMFo=", nextCursor)
	})

	s.T().Run("success with filter num cursor and seller id", func(t *testing.T) {
		orderRepo := mysql.NewOrderRepository(s.DBConn)
		orders, nextCursor, err := orderRepo.Fetch(context.TODO(), skyros.Filter{
			Num:      1,
			Cursor:   "MjAxOC0wNi0yNVQxMDowMDowMlo=",
			SellerID: "seller-id",
		})
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), orders)
		require.NotEmpty(s.T(), orders[0].Items)
		require.Equal(s.T(), mockOrder.Items[0].ProductID, orders[0].Items[0].ProductID)
		require.Equal(s.T(), mockOrder.Items[0].Quantity, orders[0].Items[0].Quantity)
		require.Equal(s.T(), "MjAxOC0wNi0yNVQxMDowMDowMFo=", nextCursor)
	})

	s.T().Run("error invalid cursor", func(t *testing.T) {
		orderRepo := mysql.NewOrderRepository(s.DBConn)
		orders, nextCursor, err := orderRepo.Fetch(context.TODO(), skyros.Filter{
			Num:    1,
			Cursor: "invalid-cursor",
		})
		require.Error(s.T(), err)
		require.Empty(s.T(), nextCursor)
		require.Empty(s.T(), orders)
	})
}

func (s *orderTestSuite) TestOrder_PatchStatus() {
	timeNow, err := time.Parse("2006-01-02T:15:04:05+07:00", "2018-06-25T:10:00:00+07:00")
	require.NoError(s.T(), err)

	var mockOrder skyros.Order
	testdata.GoldenJSONUnmarshal(s.T(), "order", &mockOrder)

	mockOrder.Seller.ID = "seller-id"
	mockOrder.Buyer.ID = "buyer-id"
	mockOrder.CreatedTime = timeNow
	mockOrder.UpdatedTime = timeNow

	s.seedOrder(mockOrder)

	s.T().Run("success", func(t *testing.T) {
		orderRepo := mysql.NewOrderRepository(s.DBConn)
		err = orderRepo.PatchStatus(context.TODO(), mockOrder.ID, 1)
		require.NoError(s.T(), err)
	})

	s.T().Run("error not found", func(t *testing.T) {
		orderRepo := mysql.NewOrderRepository(s.DBConn)
		err = orderRepo.PatchStatus(context.TODO(), "other-order-id", 1)
		require.Error(s.T(), err)
	})
}
