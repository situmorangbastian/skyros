package mysql

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/situmorangbastian/skyros"
)

type orderRepository struct {
	db *sql.DB
}

// NewOrderRepository will create the order mysql repository
func NewOrderRepository(db *sql.DB) skyros.OrderRepository {
	return orderRepository{
		db: db,
	}
}

func (r orderRepository) Store(ctx context.Context, order skyros.Order) (skyros.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return skyros.Order{}, err
	}

	timeNow := time.Now()
	order.ID = uuid.New().String()

	query, args, err := sq.Insert("orders").
		Columns("id", "buyer_id", "seller_id", "description", "source_address", "destination_address", "total_price", "status", "created_time", "updated_time").
		Values(order.ID, order.Buyer.ID, order.Seller.ID, order.Description, order.SourceAddress, order.DestinationAddress, order.TotalPrice, order.Status, timeNow, timeNow).ToSql()
	if err != nil {
		return skyros.Order{}, err
	}

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return skyros.Order{}, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Error(errors.Wrap(err, "close prepared statement"))
		}
	}()

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		return skyros.Order{}, err
	}

	for _, orderItem := range order.Items {
		query, args, err := sq.Insert("orders_product").
			Columns("id", "order_id", "product_id", "quantity", "created_time", "updated_time").
			Values(uuid.New().String(), order.ID, orderItem.ProductID, orderItem.Quantity, timeNow, timeNow).ToSql()
		if err != nil {
			return skyros.Order{}, err
		}

		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return skyros.Order{}, err
		}
		defer func() {
			if err := stmt.Close(); err != nil {
				log.Error(errors.Wrap(err, "close prepared statement"))
			}
		}()

		_, err = stmt.ExecContext(ctx, args...)
		if err != nil {
			return skyros.Order{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		return skyros.Order{}, err
	}

	return order, nil
}

func (r orderRepository) Fetch(ctx context.Context, filter skyros.Filter) ([]skyros.Order, string, error) {
	qBuilder := sq.Select("id", "buyer_id", "seller_id", "description", "source_address", "destination_address", "total_price", "created_time", "updated_time").
		From("orders").
		Where("deleted_time IS NULL").
		OrderBy("created_time DESC")

	if filter.Num > 0 {
		qBuilder = qBuilder.Limit(uint64(filter.Num))
	}

	if filter.Cursor != "" {
		decodedCursor, err := decodeCursor(filter.Cursor)
		if err != nil {
			return make([]skyros.Order, 0), "", skyros.ConstraintErrorf("invalid query param cursor")
		}
		qBuilder = qBuilder.Where(sq.Lt{"created_time": decodedCursor})
	}

	if filter.SellerID != "" {
		qBuilder = qBuilder.Where(sq.Eq{"seller_id": filter.SellerID})
	}

	query, args, err := qBuilder.ToSql()
	if err != nil {
		return []skyros.Order{}, "", err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []skyros.Order{}, "", err
	}

	orders := make([]skyros.Order, 0)
	for rows.Next() {
		order := skyros.Order{}
		err = rows.Scan(
			&order.ID,
			&order.Buyer.ID,
			&order.Seller.ID,
			&order.Description,
			&order.SourceAddress,
			&order.DestinationAddress,
			&order.TotalPrice,
			&order.CreatedTime,
			&order.UpdatedTime,
		)
		if err != nil {
			log.Error(err)
			continue
		}

		orders = append(orders, order)
	}

	errGroup := errgroup.Group{}
	for index, order := range orders {
		index, order := index, order

		order.Items = make([]skyros.OrderProduct, 0)

		errGroup.Go(func() error {
			query, args, err := sq.Select("product_id", "quantity").
				From("orders_product").
				Where(sq.Eq{"order_id": order.ID}).
				OrderBy("created_time DESC").ToSql()
			if err != nil {
				return err
			}

			rows, err := r.db.QueryContext(ctx, query, args...)
			if err != nil {
				return err
			}

			orderProducts := make([]skyros.OrderProduct, 0)
			for rows.Next() {
				orderProduct := skyros.OrderProduct{}
				err = rows.Scan(
					&orderProduct.ProductID,
					&orderProduct.Quantity,
				)
				if err != nil {
					log.Error(err)
					continue
				}

				orderProducts = append(orderProducts, orderProduct)
			}

			orders[index].Items = orderProducts
			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return []skyros.Order{}, "", err
	}

	if err = rows.Err(); err != nil {
		return []skyros.Order{}, "", err
	}

	nextCursor := ""
	if len(orders) > 0 {
		nextCursor = encodeCursor(orders[len(orders)-1].CreatedTime)
	}

	return orders, nextCursor, nil
}

func (r orderRepository) PatchStatus(ctx context.Context, ID string, status int) error {
	query, args, err := sq.Update("orders").
		Set("status", status).
		Set("updated_time", time.Now()).
		Where(sq.Eq{
			"id": ID,
		}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return skyros.ErrorNotFound("order not found")
	}

	return nil
}