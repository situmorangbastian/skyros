package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/situmorangbastian/skyros"
)

type productRepository struct {
	db *sql.DB
}

// NewProductRepository will create the product mysql repository
func NewProductRepository(db *sql.DB) skyros.ProductRepository {
	return productRepository{
		db: db,
	}
}

func (r productRepository) Store(ctx context.Context, product skyros.Product) (skyros.Product, error) {
	timeNow := time.Now()

	product.ID = uuid.New().String()

	query, args, err := sq.Insert("product").
		Columns("id", "name", "description", "price", "seller_id", "created_time", "updated_time").
		Values(product.ID, product.Name, product.Description, product.Price, product.Seller.ID, timeNow, timeNow).ToSql()
	if err != nil {
		return skyros.Product{}, err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return skyros.Product{}, err
	}

	return product, nil
}

func (r productRepository) Get(ctx context.Context, ID string) (skyros.Product, error) {
	query, args, err := sq.Select("id", "name", "description", "price", "seller_id", "created_time", "updated_time").
		From("product").
		Where(sq.Eq{"id": ID}).ToSql()
	if err != nil {
		return skyros.Product{}, err
	}

	rows := r.db.QueryRowContext(ctx, query, args...)

	product := skyros.Product{}
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
			return skyros.Product{}, skyros.ErrorNotFound("product not found")
		}
		return skyros.Product{}, err
	}

	return product, nil
}

func (r productRepository) Fetch(ctx context.Context, filter skyros.Filter) ([]skyros.Product, string, error) {
	qBuilder := sq.Select("id", "name", "description", "price", "seller_id", "created_time", "updated_time").
		From("product").
		Where("deleted_time IS NULL").
		OrderBy("created_time DESC")

	if filter.Num > 0 {
		qBuilder = qBuilder.Limit(uint64(filter.Num))
	}

	if filter.Cursor != "" {
		decodedCursor, err := decodeCursor(filter.Cursor)
		if err != nil {
			return make([]skyros.Product, 0), "", skyros.ConstraintErrorf("invalid query param cursor")
		}
		qBuilder = qBuilder.Where(sq.Lt{"created_time": decodedCursor})
	}

	if filter.Search != "" {
		keywords := strings.Split(filter.Search, ",")
		for _, keyword := range keywords {
			qBuilder = qBuilder.Where(squirrel.Like{"name": fmt.Sprintf("%%%v%%", regexp.QuoteMeta(keyword))})
		}
	}

	if filter.SellerID != "" {
		qBuilder = qBuilder.Where(sq.Eq{"seller_id": filter.SellerID})
	}

	query, args, err := qBuilder.ToSql()
	if err != nil {
		return []skyros.Product{}, "", err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []skyros.Product{}, "", err
	}

	products := make([]skyros.Product, 0)
	for rows.Next() {
		product := skyros.Product{}
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
			log.Error(err)
			continue
		}

		products = append(products, product)
	}

	nextCursor := ""
	if len(products) > 0 {
		nextCursor = encodeCursor(products[len(products)-1].CreatedTime)
	}

	return products, nextCursor, nil
}