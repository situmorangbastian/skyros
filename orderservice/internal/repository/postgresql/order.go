package postgresql

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"github.com/situmorangbastian/skyros/orderservice/internal/models"
	"github.com/situmorangbastian/skyros/orderservice/internal/repository"
	"github.com/situmorangbastian/skyros/serviceutils"
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
	defer func() { _ = tx.Rollback(ctx) }() // no-op after a successful commit

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

	limit, offset := serviceutils.PageOffset(filter.Page, filter.PageSize)
	qBuilder = qBuilder.Limit(limit).Offset(offset)

	if filter.SellerID != "" {
		qBuilder = qBuilder.Where(sq.Eq{"seller_id": filter.SellerID})
	}

	if filter.BuyerID != "" {
		qBuilder = qBuilder.Where(sq.Eq{"buyer_id": filter.BuyerID})
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
	defer rows.Close()

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
	if err = rows.Err(); err != nil {
		return []models.Order{}, err
	}
	rows.Close()

	errGroup := errgroup.Group{}
	errGroup.SetLimit(5)
	for index, order := range orders {
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
			defer rows.Close()

			orderProducts := make([]models.OrderProduct, 0)
			for rows.Next() {
				orderProduct := models.OrderProduct{}
				err = rows.Scan(
					&orderProduct.ProductID,
					&orderProduct.Quantity,
				)
				if err != nil {
					return err
				}

				orderProducts = append(orderProducts, orderProduct)
			}
			if err := rows.Err(); err != nil {
				return err
			}

			orders[index].Items = orderProducts
			return nil
		})
	}

	if err := errGroup.Wait(); err != nil {
		return []models.Order{}, err
	}

	return orders, nil
}

// PatchStatus updates the status of an order owned by sellerID.
func (r *orderRepository) PatchStatus(ctx context.Context, ID, sellerID string, status int) error {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Update("orders").
		Set("status", status).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{
			"id":        ID,
			"seller_id": sellerID,
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
