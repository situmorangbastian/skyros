package postgresql

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/situmorangbastian/skyros/orderservice/internal/models"
	"github.com/situmorangbastian/skyros/orderservice/internal/repository"
)

type orderRepository struct {
	dbpool *pgxpool.Pool
}

func NewOrderRepository(dbpool *pgxpool.Pool) repository.OrderRepository {
	return &orderRepository{
		dbpool: dbpool,
	}
}

func (r *orderRepository) Store(ctx context.Context, order models.Order) (models.Order, error) {
	tx, err := r.dbpool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return models.Order{}, err
	}

	timeNow := time.Now().UTC()
	order.ID = uuid.New().String()
	order.CreatedAt = timeNow
	order.UpdatedAt = timeNow

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Insert("orders").
		Columns(
			"id",
			"buyer_id",
			"seller_id",
			"description",
			"source_address",
			"destination_address",
			"total_price",
			"status",
			"created_at",
			"updated_at",
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
			order.CreatedAt,
			order.UpdatedAt,
		).ToSql()
	if err != nil {
		return models.Order{}, err
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return models.Order{}, err
	}

	for _, orderItem := range order.Items {
		psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
		query, args, err := psql.Insert("orders_products").
			Columns(
				"id",
				"order_id",
				"product_id",
				"quantity",
				"created_at",
				"updated_at",
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

		_, err = tx.Exec(ctx, query, args...)
		if err != nil {
			return models.Order{}, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func (r *orderRepository) Fetch(ctx context.Context, filter models.Filter) ([]models.Order, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	qBuilder := psql.Select(
		"id",
		"buyer_id",
		"seller_id",
		"description",
		"source_address",
		"destination_address",
		"total_price",
		"status",
		"created_at",
		"updated_at",
	).From("orders").OrderBy("created_at DESC")

	offset := (filter.Page - 1) * filter.PageSize
	qBuilder = qBuilder.Limit(uint64(filter.PageSize))
	if offset > 0 {
		qBuilder.Offset(uint64(offset))
	}

	if filter.SellerID != "" {
		qBuilder = qBuilder.Where(sq.Eq{"seller_id": filter.SellerID})
	}

	if filter.OrderID != "" {
		qBuilder = qBuilder.Where(sq.Eq{"id": filter.OrderID})
	}

	query, args, err := qBuilder.ToSql()
	if err != nil {
		return []models.Order{}, err
	}

	rows, err := r.dbpool.Query(ctx, query, args...)
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
			&order.CreatedAt,
			&order.UpdatedAt,
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
			psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
			query, args, err := psql.Select(
				"product_id",
				"quantity",
			).
				From("orders_products").
				Where(sq.Eq{"order_id": order.ID}).
				OrderBy("created_at DESC").ToSql()
			if err != nil {
				return err
			}

			rows, err := r.dbpool.Query(ctx, query, args...)
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
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Update("orders").
		Set("status", status).
		Set("updated_time", time.Now().UTC()).
		Where(sq.Eq{
			"id": ID,
		}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := r.dbpool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return repository.ErrNotFound
	}

	return nil
}
