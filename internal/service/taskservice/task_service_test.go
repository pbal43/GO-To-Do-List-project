package taskservice

import (
	"context"
	"errors"
	"testing"
	"toDoList/internal/domain/task/taskerrors"
	"toDoList/internal/domain/task/taskmodels"
	"toDoList/internal/server/mocks"
	"toDoList/internal/server/workers"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAllTasks(t *testing.T) {
	type want struct {
		tasks []taskmodels.Task
		err   error
	}

	tests := []struct {
		name        string
		userID      string
		dataFromDB  []taskmodels.Task
		errorFromDB error
		want        want
	}{
		{
			name:   "success",
			userID: "user1",
			dataFromDB: []taskmodels.Task{
				{
					ID:     "1",
					UserID: "user1",
					Attributes: taskmodels.TaskAttributes{
						Status:      taskmodels.StatusNew,
						Title:       "Task1",
						Description: "Desc1",
					},
				},
			},
			errorFromDB: nil,
			want: want{
				tasks: []taskmodels.Task{
					{
						ID:     "1",
						UserID: "user1",
						Attributes: taskmodels.TaskAttributes{
							Status:      taskmodels.StatusNew,
							Title:       "Task1",
							Description: "Desc1",
						},
					},
				},
				err: nil,
			},
		},
		{
			name:        "db error",
			userID:      "user2",
			dataFromDB:  []taskmodels.Task{},
			errorFromDB: errors.New("db error"),
			want: want{
				tasks: []taskmodels.Task{},
				err:   errors.New("db error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			service := NewTaskService(repo, &workers.TaskBatchDeleter{})

			repo.On("GetAllTasks", tt.userID).Return(tt.dataFromDB, tt.errorFromDB)

			got, err := service.GetAllTasks(tt.userID)
			assert.Equal(t, tt.want.tasks, got)
			assert.Equal(t, tt.want.err, err)
		})
	}
}

func TestGetTaskByID(t *testing.T) {
	type want struct {
		task taskmodels.Task
		err  error
	}

	tests := []struct {
		name        string
		taskID      string
		userID      string
		dataFromDB  taskmodels.Task
		errorFromDB error
		want        want
	}{
		{
			name:   "success",
			taskID: "1",
			userID: "u1",
			dataFromDB: taskmodels.Task{
				ID:     "1",
				UserID: "u1",
				Attributes: taskmodels.TaskAttributes{
					Status:      taskmodels.StatusNew,
					Title:       "Task1",
					Description: "Desc1",
				},
			},
			errorFromDB: nil,
			want: want{
				task: taskmodels.Task{
					ID:     "1",
					UserID: "u1",
					Attributes: taskmodels.TaskAttributes{
						Status:      taskmodels.StatusNew,
						Title:       "Task1",
						Description: "Desc1",
					},
				},
				err: nil,
			},
		},
		{
			name:        "empty taskID",
			taskID:      "",
			userID:      "u1",
			dataFromDB:  taskmodels.Task{},
			errorFromDB: nil,
			want: want{
				task: taskmodels.Task{},
				err:  taskerrors.ErrEmptyString,
			},
		},
		{
			name:        "db error",
			taskID:      "2",
			userID:      "u1",
			dataFromDB:  taskmodels.Task{},
			errorFromDB: taskerrors.ErrFoundNothing,
			want: want{
				task: taskmodels.Task{},
				err:  taskerrors.ErrFoundNothing,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			service := NewTaskService(repo, &workers.TaskBatchDeleter{})

			if tt.taskID != "" && tt.errorFromDB != nil || tt.dataFromDB.ID != "" {
				repo.On("GetTaskByID", tt.taskID, tt.userID).Return(tt.dataFromDB, tt.errorFromDB)
			}

			got, err := service.GetTaskByID(tt.taskID, tt.userID)
			assert.Equal(t, tt.want.task, got)
			assert.Equal(t, tt.want.err, err)
		})
	}
}

func TestCreateTask(t *testing.T) {
	type want struct {
		taskID string
		err    error
	}

	tests := []struct {
		name       string
		userID     string
		attributes taskmodels.TaskAttributes
		dbMock     bool
		dbErr      error
		want       want
	}{
		{
			name:   "success",
			userID: "u1",
			attributes: taskmodels.TaskAttributes{
				Status:      taskmodels.StatusNew,
				Title:       "Task1",
				Description: "Desc1",
			},
			dbMock: true,
			dbErr:  nil,
			want: want{
				taskID: "any", // We'll check only non-empty
				err:    nil,
			},
		},
		{
			name:   "invalid status",
			userID: "u1",
			attributes: taskmodels.TaskAttributes{
				Status:      "invalid",
				Title:       "Task1",
				Description: "Desc1",
			},
			dbMock: false,
			dbErr:  nil,
			want: want{
				taskID: "",
				err:    taskerrors.ErrWrongStatus,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			service := NewTaskService(repo, &workers.TaskBatchDeleter{})

			if tt.dbMock {
				repo.On("AddTask", mock.Anything).Return(tt.dbErr)
			}

			taskID, err := service.CreateTask(tt.attributes, tt.userID)

			if tt.want.err != nil {
				assert.Equal(t, tt.want.err, err)
			} else {
				assert.NotEmpty(t, taskID)
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	type want struct {
		err error
	}

	type test struct {
		name          string
		taskID        string
		userID        string
		newAttributes taskmodels.TaskAttributes
		existingTask  taskmodels.Task
		getTaskErr    error
		updateTaskErr error
		dbMockGet     bool
		dbMockUpdate  bool
		want          want
	}

	tests := []test{
		{
			name:   "success",
			taskID: "1",
			userID: "user1",
			newAttributes: taskmodels.TaskAttributes{
				Status:      taskmodels.StatusNew,
				Title:       "Updated Title",
				Description: "Updated Description",
			},
			existingTask: taskmodels.Task{
				ID:     "1",
				UserID: "user1",
				Attributes: taskmodels.TaskAttributes{
					Status:      taskmodels.StatusNew,
					Title:       "Old Title",
					Description: "Old Description",
				},
			},
			dbMockGet:    true,
			dbMockUpdate: true,
			want: want{
				err: nil,
			},
		},
		{
			name:   "invalid_status",
			taskID: "1",
			userID: "user1",
			newAttributes: taskmodels.TaskAttributes{
				Status:      "invalid_status",
				Title:       "Title",
				Description: "Description",
			},
			dbMockGet:    false,
			dbMockUpdate: false,
			want: want{
				err: taskerrors.ErrWrongStatus,
			},
		},
		{
			name:   "get_task_error",
			taskID: "1",
			userID: "user1",
			newAttributes: taskmodels.TaskAttributes{
				Status:      taskmodels.StatusNew,
				Title:       "Title",
				Description: "Description",
			},
			getTaskErr:   taskerrors.ErrDBOnGet,
			dbMockGet:    true,
			dbMockUpdate: false,
			want: want{
				err: taskerrors.ErrDBOnGet,
			},
		},
		{
			name:   "update_task_error",
			taskID: "1",
			userID: "user1",
			newAttributes: taskmodels.TaskAttributes{
				Status:      taskmodels.StatusNew,
				Title:       "Title",
				Description: "Description",
			},
			existingTask: taskmodels.Task{
				ID:     "1",
				UserID: "user1",
				Attributes: taskmodels.TaskAttributes{
					Status:      taskmodels.StatusNew,
					Title:       "Old Title",
					Description: "Old Description",
				},
			},
			dbMockGet:     true,
			dbMockUpdate:  true,
			updateTaskErr: taskerrors.ErrDBOnUpdate,
			want: want{
				err: taskerrors.ErrDBOnUpdate,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			service := NewTaskService(repo, nil)

			if tc.dbMockGet {
				repo.On("GetTaskByID", tc.taskID, tc.userID).Return(tc.existingTask, tc.getTaskErr)
			}

			if tc.dbMockUpdate {
				repo.On("UpdateTaskAttributes", mock.Anything).Return(tc.updateTaskErr)
			}

			err := service.UpdateTask(tc.taskID, tc.userID, tc.newAttributes)

			assert.Equal(t, tc.want.err, err)
		})
	}
}

func TestDeleteTaskByID(t *testing.T) {
	tests := []struct {
		name    string
		taskID  string
		userID  string
		dbErr   error
		wantErr error
	}{
		{
			name:    "success",
			taskID:  "1",
			userID:  "u1",
			dbErr:   nil,
			wantErr: nil,
		},
		{
			name:    "db error",
			taskID:  "1",
			userID:  "u1",
			dbErr:   errors.New("delete failed"),
			wantErr: errors.New("delete failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			service := NewTaskService(repo, &workers.TaskBatchDeleter{})

			repo.On("DeleteTask", tt.taskID, tt.userID).Return(tt.dbErr)

			err := service.DeleteTaskByID(tt.taskID, tt.userID)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestMarkTaskToDeleteByID(t *testing.T) {
	type want struct {
		err error
	}

	type test struct {
		name    string
		taskID  string
		userID  string
		dbError error
		want    want
	}

	tests := []test{
		{
			name:    "success",
			taskID:  "task1",
			userID:  "user1",
			dbError: nil,
			want: want{
				err: nil,
			},
		},
		{
			name:    "db error",
			taskID:  "task2",
			userID:  "user1",
			dbError: taskerrors.ErrFoundNothing,
			want: want{
				err: taskerrors.ErrFoundNothing,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			repo.On("MarkTaskToDelete", tc.taskID, tc.userID).Return(tc.dbError)

			ctx := context.Background()

			logger := log.With().Logger()
			deleter := workers.NewTaskBatchDeleter(ctx, repo, 10, logger)

			service := NewTaskService(repo, deleter)

			err := service.MarkTaskToDeleteByID(tc.taskID, tc.userID)

			assert.Equal(t, tc.want.err, err)
		})
	}
}
