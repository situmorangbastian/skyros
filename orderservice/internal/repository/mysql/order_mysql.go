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

	"github.com/situmorangbastian/skyros/orderservice/internal/domain/models"
	internalErr "github.com/situmorangbastian/skyros/orderservice/internal/errors"
	"github.com/situmorangbastian/skyros/orderservice/internal/repository"
)

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) repository.OrderRepository {
	return &orderRepository{
		db: db,
	}
}

func (r *orderRepository) Store(ctx context.Context, order models.Order) (models.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Order{}, err
	}

	timeNow := time.Now()
	order.ID = uuid.New().String()
	order.CreatedTime = timeNow
	order.UpdatedTime = timeNow

	query, args, err := sq.Insert("orders").
		Columns(
			"id",
			"buyer_id",
			"seller_id",
			"description",
			"source_address",
			"destination_address",
			"total_price",
			"status",
			"created_time",
			"updated_time",
		).
		Values(
			order.ID,
			order.Buyer.ID,
			order.Seller.ID,
			order.Description,
			order.SourceAddress,
			order.DestinationAddress,
			order.TotalPrice,
			order.Status,
			order.CreatedTime,
			order.UpdatedTime,
		).ToSql()
	if err != nil {
		return models.Order{}, err
	}

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return models.Order{}, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Error(errors.Wrap(err, "close prepared statement"))
		}
	}()

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		return models.Order{}, err
	}

	for _, orderItem := range order.Items {
		query, args, err := sq.Insert("orders_product").
			Columns(
				"id",
				"order_id",
				"product_id",
				"quantity",
				"created_time",
				"updated_time",
			).
			Values(
				uuid.New().String(),
				order.ID,
				orderItem.ProductID,
				orderItem.Quantity,
				timeNow,
				timeNow,
			).ToSql()
		if err != nil {
			return models.Order{}, err
		}

		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return models.Order{}, err
		}
		defer func() {
			if err := stmt.Close(); err != nil {
				log.Error(errors.Wrap(err, "close prepared statement"))
			}
		}()

		_, err = stmt.ExecContext(ctx, args...)
		if err != nil {
			return models.Order{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func (r *orderRepository) Fetch(ctx context.Context, filter models.Filter) ([]models.Order, error) {
	qBuilder := sq.Select(
		"id",
		"buyer_id",
		"seller_id",
		"description",
		"source_address",
		"destination_address",
		"total_price",
		"status",
		"created_time",
		"updated_time",
	).
		From("orders").
		OrderBy("created_time DESC")

	offset := (filter.Page - 1) * filter.PageSize
	qBuilder = qBuilder.Limit(uint64(filter.PageSize)).Offset(uint64(offset))

	if filter.SellerID != "" {
		qBuilder = qBuilder.Where(sq.Eq{"seller_id": filter.SellerID})
	}

	query, args, err := qBuilder.ToSql()
	if err != nil {
		return []models.Order{}, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []models.Order{}, err
	}

	orders := make([]models.Order, 0)
	for rows.Next() {
		order := models.Order{}
		err = rows.Scan(
			&order.ID,
			&order.Buyer.ID,
			&order.Seller.ID,
			&order.Description,
			&order.SourceAddress,
			&order.DestinationAddress,
			&order.TotalPrice,
			&order.Status,
			&order.CreatedTime,
			&order.UpdatedTime,
		)
		if err != nil {
			return []models.Order{}, err
		}

		orders = append(orders, order)
	}

	errGroup := errgroup.Group{}
	for index, order := range orders {
		index, order := index, order

		order.Items = make([]models.OrderProduct, 0)

		errGroup.Go(func() error {
			query, args, err := sq.Select(
				"product_id",
				"quantity",
			).
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

			orderProducts := make([]models.OrderProduct, 0)
			for rows.Next() {
				orderProduct := models.OrderProduct{}
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
		return []models.Order{}, err
	}

	if err = rows.Err(); err != nil {
		return []models.Order{}, err
	}

	return orders, nil
}

func (r *orderRepository) PatchStatus(ctx context.Context, ID string, status int) error {
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
		return internalErr.NotFoundError("order not found")
	}

	return nil
}
