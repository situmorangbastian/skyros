package postgresql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/situmorangbastian/skyros/productservice/internal/models"
	"github.com/situmorangbastian/skyros/productservice/internal/repository"
	"github.com/situmorangbastian/skyros/serviceutils"
)

// likeEscaper escapes LIKE wildcards so user input is matched literally.
var likeEscaper = strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")

type productRepository struct {
	dbpool *pgxpool.Pool
}

func NewProductRepository(dbpool *pgxpool.Pool) repository.ProductRepository {
	return &productRepository{
		dbpool: dbpool,
	}
}

func (r *productRepository) Store(ctx context.Context, product models.Product) (models.Product, error) {
	timeNow := time.Now()

	product.ID = uuid.New().String()
	product.CreatedTime = timeNow
	product.UpdatedTime = timeNow

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Insert("products").
		Columns("id", "name", "description", "price", "seller_id", "created_at", "updated_at").
		Values(product.ID, product.Name, product.Description, product.Price, product.Seller.ID, product.CreatedTime, product.UpdatedTime).ToSql()
	if err != nil {
		return models.Product{}, err
	}

	_, err = r.dbpool.Exec(ctx, query, args...)
	if err != nil {
		return models.Product{}, err
	}

	return product, nil
}

func (r *productRepository) Get(ctx context.Context, ID string) (models.Product, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Select("id", "name", "description", "price", "seller_id", "created_at", "updated_at").
		From("products").
		Where(sq.Eq{"id": ID, "deleted_at": nil}).ToSql()
	if err != nil {
		return models.Product{}, err
	}

	rows := r.dbpool.QueryRow(ctx, query, args...)
	product := models.Product{}
	err = rows.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Seller.ID,
		&product.CreatedTime,
		&product.UpdatedTime,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Product{}, status.Error(codes.NotFound, "product not found")
		}
		return models.Product{}, err
	}

	return product, nil
}

func (r *productRepository) Fetch(ctx context.Context, filter models.ProductFilter) ([]models.Product, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	qBuilder := psql.Select("id", "name", "description", "price", "seller_id", "created_at", "updated_at").
		From("products").
		Where("deleted_at IS NULL").
		OrderBy("created_at DESC")

	limit, offset := serviceutils.PageOffset(filter.Page, filter.PageSize)
	qBuilder = qBuilder.Limit(limit).Offset(offset)

	if filter.Search != "" {
		keywords := strings.Split(filter.Search, ",")
		for _, keyword := range keywords {
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				continue
			}
			qBuilder = qBuilder.Where(sq.ILike{"name": fmt.Sprintf("%%%s%%", likeEscaper.Replace(keyword))})
		}
	}

	if filter.SellerID != "" {
		qBuilder = qBuilder.Where(sq.Eq{"seller_id": filter.SellerID})
	}

	query, args, err := qBuilder.ToSql()
	if err != nil {
		return []models.Product{}, err
	}

	rows, err := r.dbpool.Query(ctx, query, args...)
	if err != nil {
		return []models.Product{}, err
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		product := models.Product{}
		err = rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Seller.ID,
			&product.CreatedTime,
			&product.UpdatedTime,
		)
		if err != nil {
			return []models.Product{}, err
		}

		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return []models.Product{}, err
	}

	return products, nil
}

func (r *productRepository) FetchByIds(ctx context.Context, ids []string) (map[string]models.Product, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	qBuilder := psql.Select("id", "name", "description", "price", "seller_id", "created_at", "updated_at").
		From("products").
		Where(sq.Eq{"id": ids, "deleted_at": nil})

	query, args, err := qBuilder.ToSql()
	if err != nil {
		return map[string]models.Product{}, err
	}

	rows, err := r.dbpool.Query(ctx, query, args...)
	if err != nil {
		return map[string]models.Product{}, err
	}
	defer rows.Close()

	products := map[string]models.Product{}
	for rows.Next() {
		product := models.Product{}
		err = rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Seller.ID,
			&product.CreatedTime,
			&product.UpdatedTime,
		)
		if err != nil {
			return map[string]models.Product{}, err
		}

		products[product.ID] = product
	}
	if err := rows.Err(); err != nil {
		return map[string]models.Product{}, err
	}

	return products, nil
}
