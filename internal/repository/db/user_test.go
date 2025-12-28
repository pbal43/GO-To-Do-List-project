package db

import (
	"context"
	"errors"
	"testing"
	"toDoList/internal/domain/user/user_errors"
	"toDoList/internal/domain/user/user_models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

// Мок для QueryRow
type mockRow struct {
	err error
}

func (r *mockRow) Scan(dest ...any) error {
	return r.err
}

// Мок для PgxIface
type mockDB struct {
	shouldDuplicate bool
}

func (m *mockDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.shouldDuplicate {
		return pgconn.NewCommandTag("INSERT"), &pgconn.PgError{
			Code:    "23505",
			Message: "duplicate key value violates unique constraint",
		}
	}
	return pgconn.NewCommandTag("INSERT"), nil
}

func (m *mockDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, errors.New("Query not implemented in mock")
}

func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return &mockRow{err: errors.New("QueryRow not implemented in mock")}
}

func (m *mockDB) Close(ctx context.Context) error {
	return nil
}

func (m *mockDB) Begin(ctx context.Context) (pgx.Tx, error) {
	return nil, nil
}

func TestUserStorage_GetAllUsers(t *testing.T) {
	tests := []struct {
		name     string
		mockRows *pgxmock.Rows
		mockErr  error
		wantLen  int
		wantErr  error
	}{
		{
			name: "two users",
			mockRows: pgxmock.NewRows([]string{"uuid", "name", "email", "password"}).
				AddRow("1", "Alice", "a@test.com", "p1").
				AddRow("2", "Bob", "b@test.com", "p2"),
			wantLen: 2,
		},
		{
			name:    "query error",
			mockErr: context.DeadlineExceeded,
			wantErr: context.DeadlineExceeded,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)

			us := &userStorage{db: mock}

			if tt.mockErr != nil {
				mock.ExpectQuery("SELECT \\* FROM users").WillReturnError(tt.mockErr)
			} else {
				mock.ExpectQuery("SELECT \\* FROM users").WillReturnRows(tt.mockRows)
			}

			users, err := us.GetAllUsers()

			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				require.Len(t, users, 0)
			} else {
				require.NoError(t, err)
				require.Len(t, users, tt.wantLen)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserStorage_GetUserByID(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		mockRows *pgxmock.Rows
		mockErr  error
		wantErr  error
		wantName string
	}{
		{
			name:     "user exists",
			userID:   "1",
			mockRows: pgxmock.NewRows([]string{"uuid", "name", "email", "password"}).AddRow("1", "Alice", "a@test.com", "p1"),
			wantName: "Alice",
		},
		{
			name:    "user not found",
			userID:  "404",
			mockErr: pgx.ErrNoRows,
			wantErr: user_errors.ErrorUserNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			us := &userStorage{db: mock}

			if tt.mockErr != nil {
				mock.ExpectQuery("SELECT \\* FROM users WHERE uuid = \\$1").WithArgs(tt.userID).WillReturnError(tt.mockErr)
			} else {
				mock.ExpectQuery("SELECT \\* FROM users WHERE uuid = \\$1").WithArgs(tt.userID).WillReturnRows(tt.mockRows)
			}

			user, err := us.GetUserByID(tt.userID)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantName, user.Name)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserStorage_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		mockRows *pgxmock.Rows
		mockErr  error
		wantErr  error
		wantName string
	}{
		{
			name:     "user exists",
			email:    "a@test.com",
			mockRows: pgxmock.NewRows([]string{"uuid", "name", "email", "password"}).AddRow("1", "Alice", "a@test.com", "p1"),
			wantName: "Alice",
		},
		{
			name:    "user not found",
			email:   "missing@test.com",
			mockErr: pgx.ErrNoRows,
			wantErr: user_errors.ErrorUserNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			us := &userStorage{db: mock}

			if tt.mockErr != nil {
				mock.ExpectQuery("SELECT \\* FROM users WHERE email = \\$1").WithArgs(tt.email).WillReturnError(tt.mockErr)
			} else {
				mock.ExpectQuery("SELECT \\* FROM users WHERE email = \\$1").WithArgs(tt.email).WillReturnRows(tt.mockRows)
			}

			user, err := us.GetUserByEmail(tt.email)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantName, user.Name)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserStorage_SaveUser(t *testing.T) {
	tests := []struct {
		name            string
		user            user_models.User
		shouldDuplicate bool
		wantErr         error
	}{
		{
			name: "success",
			user: user_models.User{"1", "Alice", "a@test.com", "p1"},
		},
		{
			name:            "duplicate",
			user:            user_models.User{"dup", "Bob", "b@test.com", "p2"},
			shouldDuplicate: true,
			wantErr:         user_errors.ErrorUserIsAlreadyExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			us := &userStorage{db: &mockDB{shouldDuplicate: tt.shouldDuplicate}}

			_, err := us.SaveUser(tt.user)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserStorage_UpdateUser(t *testing.T) {
	tests := []struct {
		name         string
		user         user_models.User
		rowsAffected int
		wantErr      error
	}{
		{
			name:         "success",
			user:         user_models.User{"1", "Alice", "a@test.com", "p1"},
			rowsAffected: 1,
		},
		{
			name:         "not found",
			user:         user_models.User{"404", "Bob", "b@test.com", "p2"},
			rowsAffected: 0,
			wantErr:      user_errors.ErrorUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			us := &userStorage{db: mock}

			mock.ExpectExec("UPDATE users").
				WithArgs(tt.user.Name, tt.user.Email, tt.user.Password, tt.user.Uuid).
				WillReturnResult(pgxmock.NewResult("UPDATE", int64(tt.rowsAffected)))

			_, err = us.UpdateUser(tt.user)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserStorage_DeleteUser(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		rowsAffected int
		wantErr      error
	}{
		{
			name:         "success",
			userID:       "1",
			rowsAffected: 1,
		},
		{
			name:         "not found",
			userID:       "404",
			rowsAffected: 0,
			wantErr:      user_errors.ErrorUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			us := &userStorage{db: mock}

			mock.ExpectExec("DELETE FROM users").
				WithArgs(tt.userID).
				WillReturnResult(pgxmock.NewResult("DELETE", int64(tt.rowsAffected)))

			err = us.DeleteUser(tt.userID)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
