package mysql

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/situmorangbastian/skyros"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository will create the user mysql repository
func NewUserRepository(db *sql.DB) skyros.UserRepository {
	return userRepository{
		db: db,
	}
}

func (r userRepository) Register(ctx context.Context, user skyros.User) (skyros.User, error) {
	timeNow := time.Now()

	user.ID = uuid.New().String()

	query, args, err := sq.Insert("user").
		Columns("id", "email", "name", "address", "password", "type", "created_time", "updated_time").
		Values(user.ID, user.Email, user.Name, user.Address, user.Password, user.Type, timeNow, timeNow).ToSql()
	if err != nil {
		return skyros.User{}, err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return skyros.User{}, err
	}

	return user, nil
}

func (r userRepository) GetUser(ctx context.Context, identifier string) (skyros.User, error) {
	query, args, err := sq.Select("id", "name", "email", "password", "address", "type").
		From("user").
		Where(sq.Or{
			sq.Eq{"email": identifier},
			sq.Eq{"id": identifier},
		}).ToSql()
	if err != nil {
		return skyros.User{}, err
	}

	rows := r.db.QueryRowContext(ctx, query, args...)

	user := skyros.User{}
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
			return skyros.User{}, skyros.ErrorNotFound("user not found")
		}
		return skyros.User{}, err
	}

	return user, nil
}
