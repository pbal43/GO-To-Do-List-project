package db

import (
	"context"
	"errors"
	"time"
	"toDoList/internal/domain/user/user_errors"
	"toDoList/internal/domain/user/user_models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type userStorage struct {
	db PgxIface
}

func (us *userStorage) GetAllUsers() ([]user_models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := us.db.Query(ctx, "SELECT * FROM users")
	if err != nil {
		return nil, err
	}

	var users []user_models.User

	for rows.Next() {
		var user user_models.User
		if err := rows.Scan(&user.Uuid, &user.Name, &user.Email, &user.Password); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (us *userStorage) GetUserByID(userID string) (user_models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user user_models.User
	err := us.db.QueryRow(ctx, "SELECT * FROM users WHERE uuid = $1", userID).
		Scan(&user.Uuid, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user_models.User{}, user_errors.ErrorUserNotExist
		}
		return user_models.User{}, err
	}

	return user, nil
}

func (us *userStorage) GetUserByEmail(email string) (user_models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user user_models.User
	err := us.db.QueryRow(ctx, "SELECT * FROM users WHERE email = $1", email).
		Scan(&user.Uuid, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user_models.User{}, user_errors.ErrorUserNotExist
		}
		return user_models.User{}, err
	}

	return user, nil
}

func (us *userStorage) SaveUser(user user_models.User) (user_models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := us.db.Exec(ctx, "INSERT INTO users (uuid, name, email, password) VALUES ($1, $2, $3, $4)",
		user.Uuid, user.Name, user.Email, user.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return user_models.User{}, user_errors.ErrorUserIsAlreadyExist
			}
		}
		return user_models.User{}, err
	}
	return user, nil
}

func (us *userStorage) UpdateUser(user user_models.User) (user_models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd, err := us.db.Exec(ctx, "UPDATE users SET name = $1, email = $2, password = $3 WHERE uuid = $4",
		user.Name, user.Email, user.Password, user.Uuid)

	if err != nil {
		return user_models.User{}, err
	}

	if cmd.RowsAffected() == 0 {
		return user_models.User{}, user_errors.ErrorUserNotFound
	}

	return user, nil
}

func (us *userStorage) DeleteUser(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd, err := us.db.Exec(ctx, "DELETE FROM users WHERE uuid = $1", userID)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return user_errors.ErrorUserNotFound
	}

	return nil
}
