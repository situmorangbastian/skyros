package mysql

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	customErrors "github.com/situmorangbastian/skyros/userservice/internal/errors"
	"github.com/situmorangbastian/skyros/userservice/internal/models"
	"github.com/situmorangbastian/skyros/userservice/internal/repository"
)

type userMysqlRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userMysqlRepo{
		db: db,
	}
}

func (r *userMysqlRepo) Register(ctx context.Context, user models.User) (models.User, error) {
	timeNow := time.Now()

	user.ID = uuid.New().String()

	query, args, err := sq.Insert("user").
		Columns(
			"id",
			"email",
			"name",
			"address",
			"password",
			"type",
			"created_time",
			"updated_time",
		).
		Values(user.ID, user.Email, user.Name, user.Address, user.Password, user.Type, timeNow, timeNow).ToSql()
	if err != nil {
		return models.User{}, err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *userMysqlRepo) GetUser(ctx context.Context, identifier string) (models.User, error) {
	query, args, err := sq.Select(
		"id",
		"name",
		"email",
		"password",
		"address",
		"type",
	).
		From("user").
		Where(sq.Or{
			sq.Eq{"email": identifier},
			sq.Eq{"id": identifier},
		}).ToSql()
	if err != nil {
		return models.User{}, err
	}

	rows := r.db.QueryRowContext(ctx, query, args...)

	user := models.User{}
	err = rows.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Address,
		&user.Type,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, customErrors.NotFoundError("user not found")
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *userMysqlRepo) FetchUsersByIDs(ctx context.Context, ids []string) (map[string]models.User, error) {
	query, args, err := sq.Select(
		"id",
		"name",
		"email",
		"address",
		"type",
	).
		From("user").
		Where(sq.Or{
			sq.Eq{"email": ids},
			sq.Eq{"id": ids},
		}).ToSql()
	if err != nil {
		return map[string]models.User{}, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return map[string]models.User{}, err
	}

	users := map[string]models.User{}
	for rows.Next() {
		user := models.User{}
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Address,
			&user.Type,
		)
		if err != nil {
			return map[string]models.User{}, err
		}

		users[user.ID] = user
	}

	return users, nil
}
