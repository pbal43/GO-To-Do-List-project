package db

import (
	"errors"
	"testing"
	"toDoList/internal/domain/task/task_errors"
	"toDoList/internal/domain/task/task_models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func TestTaskStorage_AddTask(t *testing.T) {
	tests := []struct {
		name            string
		task            task_models.Task
		shouldDuplicate bool
		wantErr         error
	}{
		{
			name: "success",
			task: task_models.Task{
				ID:     "1",
				UserID: "u1",
				Attributes: task_models.TaskAttributes{
					Status:      task_models.StatusNew,
					Title:       "t1",
					Description: "d1",
				},
			},
		},
		{
			name: "duplicate",
			task: task_models.Task{
				ID:     "2",
				UserID: "u2",
				Attributes: task_models.TaskAttributes{
					Status:      task_models.StatusInProgress,
					Title:       "t2",
					Description: "d2",
				},
			},
			shouldDuplicate: true,
			wantErr:         task_errors.ErrorTaskIsAlreadyExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			ts := &taskStorage{db: mock}

			exec := mock.ExpectExec("INSERT INTO tasks").
				WithArgs(tt.task.ID, tt.task.UserID, tt.task.Attributes.Status, tt.task.Attributes.Title, tt.task.Attributes.Description)

			if tt.shouldDuplicate {
				exec.WillReturnError(&pgconn.PgError{Code: "23505"})
			} else {
				exec.WillReturnResult(pgxmock.NewResult("INSERT", 1))
			}

			err = ts.AddTask(tt.task)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskStorage_UpdateTaskAttributes(t *testing.T) {
	tests := []struct {
		name         string
		task         task_models.Task
		rowsAffected int64
		wantErr      error
	}{
		{
			"success",
			task_models.Task{
				ID: "1",
				Attributes: task_models.TaskAttributes{
					Status:      task_models.StatusNew,
					Title:       "t1",
					Description: "d1",
				},
			},
			1,
			nil,
		},
		{
			"not found",
			task_models.Task{ID: "404"},
			0,
			task_errors.FoundNothingErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			ts := &taskStorage{db: mock}

			mock.ExpectExec("UPDATE tasks").
				WithArgs(tt.task.Attributes.Status, tt.task.Attributes.Title, tt.task.Attributes.Description, tt.task.ID).
				WillReturnResult(pgxmock.NewResult("UPDATE", tt.rowsAffected))

			err = ts.UpdateTaskAttributes(tt.task)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskStorage_DeleteTask(t *testing.T) {
	tests := []struct {
		name         string
		taskID       string
		userID       string
		rowsAffected int64
		wantErr      error
	}{
		{"success", "1", "u1", 1, nil},
		{"not found", "404", "u2", 0, task_errors.FoundNothingErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			ts := &taskStorage{db: mock}

			mock.ExpectExec("DELETE FROM tasks").
				WithArgs(tt.taskID, tt.userID).
				WillReturnResult(pgxmock.NewResult("DELETE", tt.rowsAffected))

			err = ts.DeleteTask(tt.taskID, tt.userID)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskStorage_MarkTaskToDelete(t *testing.T) {
	tests := []struct {
		name         string
		taskID       string
		userID       string
		rowsAffected int64
		wantErr      error
	}{
		{"success", "1", "u1", 1, nil},
		{"not found", "404", "u2", 0, task_errors.FoundNothingErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			ts := &taskStorage{db: mock}

			mock.ExpectExec("UPDATE tasks SET deleted = true").
				WithArgs(tt.taskID, tt.userID).
				WillReturnResult(pgxmock.NewResult("UPDATE", tt.rowsAffected))

			err = ts.MarkTaskToDelete(tt.taskID, tt.userID)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskStorage_DeleteMarkedTasks(t *testing.T) {
	tests := []struct {
		name       string
		prepareErr error
		execErr    error
		commitErr  error
		wantErr    error
	}{
		{"success", nil, nil, nil, nil},
		{"prepare_error", errors.New("prepare failed"), nil, nil, errors.New("prepare failed")},
		{"exec_error", nil, errors.New("exec failed"), nil, errors.New("exec failed")},
		{"commit_error", nil, nil, errors.New("commit failed"), errors.New("commit failed")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			ts := &taskStorage{db: mock}

			mock.ExpectBegin()

			if tt.prepareErr != nil {
				mock.ExpectPrepare("delete_tasks", "DELETE FROM tasks WHERE deleted = true").
					WillReturnError(tt.prepareErr)
				mock.ExpectRollback()
			} else {
				mock.ExpectPrepare("delete_tasks", "DELETE FROM tasks WHERE deleted = true")

				if tt.execErr != nil {
					mock.ExpectExec("delete_tasks").WillReturnError(tt.execErr)
					mock.ExpectRollback()
				} else {
					mock.ExpectExec("delete_tasks").WillReturnResult(pgxmock.NewResult("DELETE", 2))

					if tt.commitErr != nil {
						mock.ExpectCommit().WillReturnError(tt.commitErr)
					} else {
						mock.ExpectCommit()
					}
				}
			}

			err = ts.DeleteMarkedTasks()
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskStorage_GetAllTasks(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		mockData []task_models.Task
		mockErr  error
		wantErr  bool
	}{
		{
			name:   "success",
			userID: "user1",
			mockData: []task_models.Task{
				{ID: "1", UserID: "user1", Attributes: task_models.TaskAttributes{Status: "New", Title: "t1", Description: "d1"}},
				{ID: "2", UserID: "user1", Attributes: task_models.TaskAttributes{Status: "Done", Title: "t2", Description: "d2"}},
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "query_error",
			userID:  "user2",
			mockErr: errors.New("query failed"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			ts := &taskStorage{db: mock}

			if tt.mockErr != nil {
				mock.ExpectQuery("SELECT \\* FROM tasks where userid = \\$1").
					WithArgs(tt.userID).
					WillReturnError(tt.mockErr)
			} else {
				rows := pgxmock.NewRows([]string{"id", "userid", "status", "title", "description", "deleted"})
				for _, task := range tt.mockData {
					rows.AddRow(task.ID, task.UserID, task.Attributes.Status, task.Attributes.Title, task.Attributes.Description, task.Deleted)
				}
				mock.ExpectQuery("SELECT \\* FROM tasks where userid = \\$1").
					WithArgs(tt.userID).
					WillReturnRows(rows)
			}

			got, err := ts.GetAllTasks(tt.userID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.mockData, got)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskStorage_GetTaskByID(t *testing.T) {
	tests := []struct {
		name     string
		taskID   string
		userID   string
		mockData task_models.Task
		mockErr  error
		wantErr  error
	}{
		{
			name:   "success",
			taskID: "1",
			userID: "user1",
			mockData: task_models.Task{
				ID: "1", UserID: "user1", Attributes: task_models.TaskAttributes{Status: "New", Title: "t1", Description: "d1"},
			},
			wantErr: nil,
		},
		{
			name:    "not_found",
			taskID:  "2",
			userID:  "user1",
			mockErr: pgx.ErrNoRows,
			wantErr: task_errors.FoundNothingErr,
		},
		{
			name:    "query_error",
			taskID:  "3",
			userID:  "user1",
			mockErr: errors.New("query failed"),
			wantErr: errors.New("query failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewConn()
			require.NoError(t, err)
			ts := &taskStorage{db: mock}

			if tt.mockErr != nil {
				mock.ExpectQuery("SELECT \\* FROM tasks WHERE id = \\$1 AND userid = \\$2").
					WithArgs(tt.taskID, tt.userID).
					WillReturnError(tt.mockErr)
			} else {
				rows := pgxmock.NewRows([]string{"id", "userid", "status", "title", "description", "deleted"}).
					AddRow(tt.mockData.ID, tt.mockData.UserID, tt.mockData.Attributes.Status,
						tt.mockData.Attributes.Title, tt.mockData.Attributes.Description, tt.mockData.Deleted)

				mock.ExpectQuery("SELECT \\* FROM tasks WHERE id = \\$1 AND userid = \\$2").
					WithArgs(tt.taskID, tt.userID).
					WillReturnRows(rows)
			}

			got, err := ts.GetTaskByID(tt.taskID, tt.userID)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.mockData, got)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
