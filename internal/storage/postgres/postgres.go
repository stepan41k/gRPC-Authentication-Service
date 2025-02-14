package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stepan41k/gRPC/internal/domain/models"
	"github.com/stepan41k/gRPC/internal/storage"
)

type Storage struct {
	pool *pgxpool.Pool
	mx sync.Mutex
}
func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.postgres.New"

	pool, err := pgxpool.Connect(context.Background(), storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w",op ,err)
	}

	return &Storage{pool: pool, mx: sync.Mutex{}}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (id int64, err error) {
	const op = "storage.postgres.postgres.SaveUser"

	tx, err := s.pool.Begin(ctx)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}

		defer func() {
			if err != nil {
				_ = tx.Rollback(context.Background())
				return
			}

			commitErr := tx.Commit(ctx)
			if commitErr != nil {
				err = fmt.Errorf("%s: %w", op, err)
			}
		}()

		err = tx.QueryRow(ctx, `
			INSERT INTO users (email, pass_hash)
			VALUES($1, $2)
			RETURNING id;
		`, email, passHash).Scan(&id)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, nil
			}

			return 0, fmt.Errorf("%s: %w", op, err)
		}

		return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.postgres.User"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}

		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			err = fmt.Errorf("%s: %w", op, err)
		}
	}()

	row := tx.QueryRow(ctx, `
		SELECT id, email, pass_hash
		FROM users
		WHERE email = $1 
	`, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.postgres.postgres.IsAdmin"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("%s: %w",op, err)
	}

	defer func()  {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}	

		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			err = fmt.Errorf("%s: %w", op, err)
		}
	}()

	var isAdmin int64
	row := tx.QueryRow(ctx, `
		SELECT admins.id
		FROM admins
		WHERE admins.user_id = $1;
	`, userID)

	err = row.Scan(&isAdmin)
	// if isAdmin == 0 {
	// 	return false, nil
	// }
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, err)
		}

		return false, fmt.Errorf("%s: %w",op ,err)
	}

	return true, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.postgres.postgres.App"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}

		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			err = fmt.Errorf("%s: %w", op, err)
		}
	}()

	var app models.App
	row := tx.QueryRow(ctx, `
		SELECT id, name, secret
		FROM apps
		WHERE id = $1
	`, appID)

	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s, %w", op, err)
	}

	return app, nil
}

