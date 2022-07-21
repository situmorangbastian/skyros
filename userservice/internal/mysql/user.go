package mysql

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/situmorangbastian/skyros/userservice"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) userservice.UserRepository {
	return userRepository{
		db: db,
	}
}

func (r userRepository) Register(ctx context.Context, user userservice.User) (userservice.User, error) {
	timeNow := time.Now()

	user.ID = uuid.New().String()

	query, args, err := sq.Insert("user").
		Columns("id", "email", "name", "address", "password", "type", "created_time", "updated_time").
		Values(user.ID, user.Email, user.Name, user.Address, user.Password, user.Type, timeNow, timeNow).ToSql()
	if err != nil {
		return userservice.User{}, err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return userservice.User{}, err
	}

	return user, nil
}

func (r userRepository) GetUser(ctx context.Context, identifier string) (userservice.User, error) {
	query, args, err := sq.Select("id", "name", "email", "password", "address", "type").
		From("user").
		Where(sq.Or{
			sq.Eq{"email": identifier},
			sq.Eq{"id": identifier},
		}).ToSql()
	if err != nil {
		return userservice.User{}, err
	}

	rows := r.db.QueryRowContext(ctx, query, args...)

	user := userservice.User{}
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
			return userservice.User{}, userservice.ErrorNotFound("user not found")
		}
		return userservice.User{}, err
	}

	return user, nil
}

func (r userRepository) FetchUsersByIDs(ctx context.Context, ids []string) (map[string]userservice.User, error) {
	query, args, err := sq.Select("id", "name", "email", "address", "type").
		From("user").
		Where(sq.Or{
			sq.Eq{"email": ids},
			sq.Eq{"id": ids},
		}).ToSql()
	if err != nil {
		return map[string]userservice.User{}, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return map[string]userservice.User{}, err
	}

	users := map[string]userservice.User{}
	for rows.Next() {
		user := userservice.User{}
		err = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Address,
			&user.Type,
		)
		if err != nil {
			return map[string]userservice.User{}, err
		}

		users[user.ID] = user
	}

	return users, nil
}
