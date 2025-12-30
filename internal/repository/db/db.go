package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" //
	"github.com/rs/zerolog/log"
)

type Storage struct {
	userStorage
	taskStorage
}

// PgxIface - общий интерфейс для мока/адаптера.
type PgxIface interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Close(ctx context.Context) error
	Begin(ctx context.Context) (pgx.Tx, error)
}

// pgxConnAdapter - адаптер для *pgx.Conn.
type pgxConnAdapter struct {
	*pgx.Conn
}

func (a pgxConnAdapter) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return a.Conn.Exec(ctx, sql, args...)
}

func (a pgxConnAdapter) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return a.Conn.Query(ctx, sql, args...)
}

func (a pgxConnAdapter) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return a.Conn.QueryRow(ctx, sql, args...)
}

func (a pgxConnAdapter) Close(ctx context.Context) error {
	return a.Conn.Close(ctx)
}

func NewStorage(connStr string) (*Storage, error) {
	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	adapter := pgxConnAdapter{Conn: db}

	return &Storage{
		userStorage: userStorage{db: adapter},
		taskStorage: taskStorage{db: adapter},
	}, nil
}

func (s *Storage) Close(ctx context.Context) error {
	return s.userStorage.db.Close(ctx)
}

func Migrations(dsn string, migratePath string) error {
	mPath := fmt.Sprintf("file://%s", migratePath)
	m, err := migrate.New(mPath, dsn)

	if err != nil {
		return err
	}

	if err = m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		log.Printf("DB is already up to date")
	}

	log.Printf("Migration complete")

	return nil
}
