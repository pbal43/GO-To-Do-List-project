package db

import (
	"context"
	"errors"
	"toDoList/internal"
	"toDoList/internal/domain/user/usererrors"
	"toDoList/internal/domain/user/usermodels"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type userStorage struct {
	db PgxIface
}

func (us *userStorage) GetAllUsers() ([]usermodels.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	rows, err := us.db.Query(ctx, "SELECT uuid, name, email, password FROM users")
	if err != nil {
		return nil, err
	}

	var users []usermodels.User

	for rows.Next() {
		var user usermodels.User
		if err = rows.Scan(&user.UUID, &user.Name, &user.Email, &user.Password); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (us *userStorage) GetUserByID(userID string) (usermodels.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	var user usermodels.User
	err := us.db.QueryRow(ctx, "SELECT uuid, name, email, password FROM users WHERE uuid = $1", userID).
		Scan(&user.UUID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usermodels.User{}, usererrors.ErrUserNotExist
		}
		return usermodels.User{}, err
	}

	return user, nil
}

func (us *userStorage) GetUserByEmail(email string) (usermodels.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	var user usermodels.User
	err := us.db.QueryRow(ctx, "SELECT uuid, name, email, password FROM users WHERE email = $1", email).
		Scan(&user.UUID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usermodels.User{}, usererrors.ErrUserNotExist
		}
		return usermodels.User{}, err
	}

	return user, nil
}

func (us *userStorage) SaveUser(user usermodels.User) (usermodels.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	_, err := us.db.Exec(ctx, "INSERT INTO users (uuid, name, email, password) VALUES ($1, $2, $3, $4)",
		user.UUID, user.Name, user.Email, user.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return usermodels.User{}, usererrors.ErrUserIsAlreadyExist
			}
		}
		return usermodels.User{}, err
	}
	return user, nil
}

func (us *userStorage) UpdateUser(user usermodels.User) (usermodels.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	cmd, err := us.db.Exec(ctx, "UPDATE users SET name = $1, email = $2, password = $3 WHERE uuid = $4",
		user.Name, user.Email, user.Password, user.UUID)

	if err != nil {
		return usermodels.User{}, err
	}

	if cmd.RowsAffected() == 0 {
		return usermodels.User{}, usererrors.ErrUserNotFound
	}

	return user, nil
}

func (us *userStorage) DeleteUser(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	cmd, err := us.db.Exec(ctx, "DELETE FROM users WHERE uuid = $1", userID)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return usererrors.ErrUserNotFound
	}

	return nil
}
