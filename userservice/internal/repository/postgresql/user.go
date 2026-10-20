package postgresql

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/situmorangbastian/skyros/userservice/internal/models"
	"github.com/situmorangbastian/skyros/userservice/internal/repository"
)

type userRepo struct {
	dbpool *pgxpool.Pool
}

func NewUserRepository(dbpool *pgxpool.Pool) repository.UserRepository {
	return &userRepo{
		dbpool: dbpool,
	}
}

func (r *userRepo) Register(ctx context.Context, user models.User) (models.User, error) {
	timeNow := time.Now()
	user.ID = uuid.New().String()
	userData, _ := json.Marshal(user.Data)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Insert("users").
		Columns(
			"id",
			"email",
			"name",
			"password",
			"user_data",
			"created_at",
			"updated_at",
		).
		Values(user.ID, user.Email, user.Name, user.Password, userData, timeNow, timeNow).ToSql()
	if err != nil {
		return models.User{}, err
	}

	_, err = r.dbpool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return models.User{}, repository.ErrAlreadyExists
		}
		return models.User{}, err
	}
	return user, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Select(
		"id",
		"email",
		"name",
		"password",
		"user_data",
	).
		From("users").
		Where(sq.Or{
			sq.Eq{"email": email},
		}).ToSql()
	if err != nil {
		return models.User{}, err
	}

	rows := r.dbpool.QueryRow(ctx, query, args...)

	var userData []byte

	user := models.User{}
	err = rows.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&userData,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, repository.ErrNotFound
		}
		return models.User{}, err
	}
	err = json.Unmarshal(userData, &user.Data)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *userRepo) FetchUsersByIDs(ctx context.Context, ids []string) (map[string]models.User, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Select(
		"id",
		"email",
		"name",
		"user_data",
	).
		From("users").
		Where(sq.Or{
			sq.Eq{"id": ids},
		}).ToSql()
	if err != nil {
		return map[string]models.User{}, err
	}

	rows, err := r.dbpool.Query(ctx, query, args...)
	if err != nil {
		return map[string]models.User{}, err
	}
	defer rows.Close()

	users := map[string]models.User{}
	for rows.Next() {
		user := models.User{}
		var userData []byte
		err = rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&userData,
		)
		if err != nil {
			return map[string]models.User{}, err
		}
		err = json.Unmarshal(userData, &user.Data)
		if err != nil {
			return map[string]models.User{}, err
		}
		users[user.ID] = user
	}
	if err := rows.Err(); err != nil {
		return map[string]models.User{}, err
	}
	return users, nil
}
