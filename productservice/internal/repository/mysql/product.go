package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/situmorangbastian/skyros/productservice/internal/models"
	"github.com/situmorangbastian/skyros/productservice/internal/repository"
)

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) repository.ProductRepository {
	return productRepository{
		db: db,
	}
}

func (r productRepository) Store(ctx context.Context, product models.Product) (models.Product, error) {
	timeNow := time.Now()

	product.ID = uuid.New().String()
	product.CreatedTime = timeNow
	product.UpdatedTime = timeNow

	query, args, err := sq.Insert("product").
		Columns("id", "name", "description", "price", "seller_id", "created_time", "updated_time").
		Values(product.ID, product.Name, product.Description, product.Price, product.Seller.ID, product.CreatedTime, product.UpdatedTime).ToSql()
	if err != nil {
		return models.Product{}, err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return models.Product{}, err
	}

	return product, nil
}

func (r productRepository) Get(ctx context.Context, ID string) (models.Product, error) {
	query, args, err := sq.Select("id", "name", "description", "price", "seller_id", "created_time", "updated_time").
		From("product").
		Where(sq.Eq{"id": ID}).ToSql()
	if err != nil {
		return models.Product{}, err
	}

	rows := r.db.QueryRowContext(ctx, query, args...)

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
		if err == sql.ErrNoRows {
			return models.Product{}, status.Error(codes.NotFound, "product not found")
		}
		return models.Product{}, err
	}

	return product, nil
}

func (r productRepository) Fetch(ctx context.Context, filter models.ProductFilter) ([]models.Product, error) {
	qBuilder := sq.Select("id", "name", "description", "price", "seller_id", "created_time", "updated_time").
		From("product").
		Where("deleted_time IS NULL").
		OrderBy("created_time DESC")

	offset := (filter.Page - 1) * filter.PageSize
	qBuilder = qBuilder.Limit(uint64(filter.PageSize))
	if offset > 0 {
		qBuilder.Offset(uint64(offset))
	}

	if filter.Search != "" {
		keywords := strings.Split(filter.Search, ",")
		for _, keyword := range keywords {
			qBuilder = qBuilder.Where(sq.Like{"name": fmt.Sprintf("%%%v%%", regexp.QuoteMeta(keyword))})
		}
	}

	if filter.SellerID != "" {
		qBuilder = qBuilder.Where(sq.Eq{"seller_id": filter.SellerID})
	}

	query, args, err := qBuilder.ToSql()
	if err != nil {
		return []models.Product{}, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []models.Product{}, err
	}

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

	return products, nil
}

func (r productRepository) FetchByIds(ctx context.Context, ids []string) (map[string]models.Product, error) {
	qBuilder := sq.Select("id", "name", "description", "price", "seller_id", "created_time", "updated_time").
		From("product").
		Where(sq.Eq{"id": ids})

	query, args, err := qBuilder.ToSql()
	if err != nil {
		return map[string]models.Product{}, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return map[string]models.Product{}, err
	}

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

	return products, nil
}
