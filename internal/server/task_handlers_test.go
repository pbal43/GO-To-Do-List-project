package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"toDoList/internal/domain/task/taskmodels"
	"toDoList/internal/server/mocks"
	"toDoList/internal/server/workers"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTasks(t *testing.T) {
	var srv ToDoListAPI
	gin.SetMode(gin.ReleaseMode)

	testTasks := []taskmodels.Task{
		{ID: "task1", UserID: "user1", Attributes: taskmodels.TaskAttributes{Title: "Task 1"}},
		{ID: "task2", UserID: "user1", Attributes: taskmodels.TaskAttributes{Title: "Task 2"}},
	}

	tests := []struct {
		name         string
		userIDCtx    any
		mockFlag     bool
		mockTasks    []taskmodels.Task
		mockErr      error
		wantStatus   int
		wantContains string
	}{
		{
			name:         "Get tasks success",
			userIDCtx:    "user1",
			mockFlag:     true,
			mockTasks:    testTasks,
			mockErr:      nil,
			wantStatus:   http.StatusOK,
			wantContains: "Task 1",
		},
		{
			name:         "Empty task list",
			userIDCtx:    "user1",
			mockFlag:     true,
			mockTasks:    []taskmodels.Task{},
			mockErr:      nil,
			wantStatus:   http.StatusOK,
			wantContains: "Task list is empty",
		},
		{
			name:         "Unauthorized",
			mockFlag:     false,
			wantStatus:   http.StatusUnauthorized,
			wantContains: "unauthorized",
		},
		{
			name:         "Wrong user id type",
			userIDCtx:    123,
			mockFlag:     false,
			wantStatus:   http.StatusInternalServerError,
			wantContains: "userID has wrong type",
		},
		{
			name:         "DB error",
			userIDCtx:    "user1",
			mockFlag:     true,
			mockErr:      errors.New("db error"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			srv.db = repo

			taskDeleter := workers.NewTaskBatchDeleter(context.Background(), srv.db, 10, zerolog.Nop())
			srv.taskDeleter = taskDeleter

			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.userIDCtx != nil {
					c.Set("userID", tc.userIDCtx)
				}
				c.Next()
			})

			r.GET("/tasks", srv.getTasks)

			if tc.mockFlag {
				repo.On("GetAllTasks", tc.userIDCtx).Return(tc.mockTasks, tc.mockErr)
			}

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req := resty.New().R().SetHeader("Content-Type", "application/json")
			res, err := req.Get(httpSrv.URL + "/tasks")

			assert.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.StatusCode())
			assert.Contains(t, string(res.Body()), tc.wantContains)
		})
	}
}

func TestCreateTask(t *testing.T) {
	var srv ToDoListAPI
	gin.SetMode(gin.ReleaseMode)

	type want struct {
		body       string
		statusCode int
	}

	type test struct {
		name       string
		taskJSON   string
		taskFromDB taskmodels.Task
		userIDCtx  any
		mockFlag   bool
		err        error
		want       want
	}

	tests := []test{
		{
			name:      "Create_task_success",
			userIDCtx: "user123",
			taskJSON: `{
				"title":"New Task",
				"description":"Task description",
				"status":"New"
			}`,
			taskFromDB: taskmodels.Task{
				ID: "task123",
				Attributes: taskmodels.TaskAttributes{
					Title:       "New Task",
					Description: "Task description",
					Status:      "New",
				},
				UserID: "user123",
			},
			mockFlag: true,
			err:      nil,
			want: want{
				body:       `{"TaskID":`,
				statusCode: http.StatusOK,
			},
		},
		{
			name:      "Bad_request_missing_status",
			userIDCtx: "user1",
			taskJSON: `{
				"title":"Bad Task",
				"description":"Missing status"
			}`,
			mockFlag: false,
			err:      nil,
			want: want{
				body:       "Key: 'TaskAttributes.Status' Error:Field validation for 'Status' failed",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:      "Bad_request_wrong_status",
			userIDCtx: "user1",
			taskJSON: `{
				"title":"Bad Task",
				"description":"Wrong status",
				"status":"Unknown"
			}`,
			mockFlag: false,
			err:      nil,
			want: want{
				body:       "wrong status",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "Unauthorized",
			taskJSON: `{
				"title":"Bad Task",
				"description":"Wrong status",
				"status":"Unknown"
			}`,
			mockFlag: false,
			err:      nil,
			want: want{
				body:       "unauthorized",
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:      "Wrong user id type",
			userIDCtx: 123,
			taskJSON: `{
				"title":"Bad Task",
				"description":"Wrong status",
				"status":"Unknown"
			}`,
			mockFlag: false,
			err:      nil,
			want: want{
				body:       "userID has wrong type",
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.userIDCtx != nil {
					c.Set("userID", tc.userIDCtx)
				}
				c.Next()
			})

			r.POST("/tasks", srv.createTask)
			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			repo := mocks.NewStorage(t)
			srv.db = repo

			taskDeleter := workers.NewTaskBatchDeleter(context.Background(), srv.db, 10, zerolog.Nop())
			srv.taskDeleter = taskDeleter

			if tc.mockFlag {
				repo.On("AddTask", mock.MatchedBy(func(task taskmodels.Task) bool {
					return task.Attributes.Title == tc.taskFromDB.Attributes.Title &&
						task.Attributes.Description == tc.taskFromDB.Attributes.Description &&
						task.Attributes.Status == tc.taskFromDB.Attributes.Status &&
						task.UserID == tc.taskFromDB.UserID
				})).Return(tc.err).Run(func(args mock.Arguments) {
					argTask := args.Get(0).(taskmodels.Task)
					argTask.ID = tc.taskFromDB.ID //nolint:govet // Ругается, хотя нужно для тестов, оставляем
				})
			}

			req := resty.New().R()
			req.URL = httpSrv.URL + "/tasks"
			req.Method = http.MethodPost
			req.Body = tc.taskJSON

			res, err := req.Send()
			assert.NoError(t, err)
			assert.Equal(t, tc.want.statusCode, res.StatusCode())
			assert.Contains(t, string(res.Body()), tc.want.body)
		})
	}
}

func TestGetTaskByID(t *testing.T) {
	var srv ToDoListAPI
	gin.SetMode(gin.ReleaseMode)

	type want struct {
		body       string
		statusCode int
	}

	type test struct {
		name       string
		taskID     string
		taskFromDB taskmodels.Task
		mockFlag   bool
		err        error
		userIDCtx  any
		want       want
	}

	tests := []test{
		{
			name:      "Get_task_success",
			taskID:    "task123",
			userIDCtx: "user123",
			taskFromDB: taskmodels.Task{
				ID:     "task123",
				UserID: "user123",
				Attributes: taskmodels.TaskAttributes{
					Title:       "Test task",
					Description: "Desc",
					Status:      taskmodels.StatusNew,
				},
			},
			mockFlag: true,
			err:      nil,
			want: want{
				body:       `"id":"task123"`,
				statusCode: http.StatusOK,
			},
		},
		{
			name:      "Task_not_found",
			userIDCtx: "user123",
			taskID:    "task404",
			mockFlag:  true,
			err:       errors.New("task not found"),
			want: want{
				body:       "task not found",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:     "Unauthorized",
			taskID:   "task404",
			mockFlag: false,
			want: want{
				body:       "unauthorized",
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:      "Wrong user id type",
			taskID:    "task404",
			userIDCtx: 123,
			mockFlag:  false,
			want: want{
				body:       "userID has wrong type",
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			srv.db = repo
			r := gin.New()

			r.Use(func(c *gin.Context) {
				if tc.userIDCtx != nil {
					c.Set("userID", tc.userIDCtx)
				}
				c.Next()
			})

			r.GET("/tasks/:id", srv.getTaskByID)
			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			taskDeleter := workers.NewTaskBatchDeleter(context.Background(), srv.db, 10, zerolog.Nop())
			srv.taskDeleter = taskDeleter

			if tc.mockFlag {
				repo.On("GetTaskByID", tc.taskID, "user123").
					Return(tc.taskFromDB, tc.err)
			}

			req := resty.New().R()
			req.URL = httpSrv.URL + "/tasks/" + tc.taskID
			req.Method = http.MethodGet

			res, err := req.Send()
			assert.NoError(t, err)
			assert.Equal(t, tc.want.statusCode, res.StatusCode())
			assert.Contains(t, string(res.Body()), tc.want.body)
		})
	}
}

func TestUpdateTask(t *testing.T) {
	var srv ToDoListAPI
	gin.SetMode(gin.ReleaseMode)

	type want struct {
		body       string
		statusCode int
	}

	type test struct {
		name      string
		taskID    string
		userIDCtx any
		bodyJSON  string
		mockFlag  bool
		err       error
		want      want
	}

	tests := []test{
		{
			name:      "Update_task_success",
			taskID:    "task123",
			userIDCtx: "user123",
			bodyJSON: `{
				"title": "Updated",
				"description": "Updated desc",
				"status": "In Progress"
			}`,
			mockFlag: true,
			err:      nil,
			want: want{
				body:       "TaskID: task123 was updated",
				statusCode: http.StatusOK,
			},
		},
		{
			name:      "Bad_request_wrong_status",
			taskID:    "task123",
			userIDCtx: "user123",
			bodyJSON: `{
				"title": "Updated",
				"description": "Updated desc",
				"status": "Wrong"
			}`,
			mockFlag: false,
			want: want{
				body:       "wrong status",
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name:      "Bad_request_invalid_json",
			taskID:    "task123",
			userIDCtx: "user123",
			bodyJSON: `{
				"title": "Updated
			}`,
			mockFlag: false,
			want: want{
				body:       "invalid character",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "Unathorized",
			taskID: "task123",
			bodyJSON: `{
				"title": "Updated",
				"description": "Updated desc",
				"status": "In Progress"
			}`,
			mockFlag: false,
			want: want{
				body:       "unauthorized",
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:      "Bad_request_invalid_json",
			taskID:    "task123",
			userIDCtx: 123,
			bodyJSON: `{
				"title": "Updated",
				"description": "Updated desc",
				"status": "In Progress"
			}`,
			mockFlag: false,
			want: want{
				body:       "userID has wrong type",
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.userIDCtx != nil {
					c.Set("userID", tc.userIDCtx)
				}
				c.Next()
			})

			r.PUT("/tasks/:id", srv.updateTask)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			repo := mocks.NewStorage(t)
			srv.db = repo

			taskDeleter := workers.NewTaskBatchDeleter(context.Background(), repo, 10, zerolog.Nop())
			srv.taskDeleter = taskDeleter

			if tc.mockFlag {
				repo.On(
					"GetTaskByID",
					tc.taskID,
					"user123",
				).Return(taskmodels.Task{
					ID:     tc.taskID,
					UserID: "user123",
					Attributes: taskmodels.TaskAttributes{
						Title:       "Old",
						Description: "Old",
						Status:      "New",
					},
				}, nil)

				repo.On(
					"UpdateTaskAttributes",
					mock.MatchedBy(func(task taskmodels.Task) bool {
						return task.ID == tc.taskID &&
							task.UserID == "user123" &&
							task.Attributes.Title == "Updated" &&
							task.Attributes.Status == "In Progress"
					}),
				).Return(tc.err)
			}

			req := resty.New().R()
			req.URL = httpSrv.URL + "/tasks/" + tc.taskID
			req.Method = http.MethodPut
			req.Body = tc.bodyJSON

			res, err := req.Send()
			assert.NoError(t, err)
			assert.Equal(t, tc.want.statusCode, res.StatusCode())
			assert.Contains(t, string(res.Body()), tc.want.body)
		})
	}
}

func TestDeleteTask(t *testing.T) {
	var srv ToDoListAPI
	gin.SetMode(gin.ReleaseMode)

	type want struct {
		body       string
		statusCode int
	}

	type test struct {
		name      string
		userIDCtx any
		taskID    string
		mockFlag  bool
		err       error
		want      want
	}

	tests := []test{
		{
			name:      "Delete_task_success",
			taskID:    "task123",
			userIDCtx: "user123",
			mockFlag:  true,
			err:       nil,
			want: want{
				body:       "Task was deleted",
				statusCode: http.StatusOK,
			},
		},
		{
			name:      "Delete_task_error",
			taskID:    "task123",
			userIDCtx: "user123",
			mockFlag:  true,
			err:       errors.New("delete failed"),
			want: want{
				body:       "delete failed",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:     "Unathorized",
			taskID:   "task123",
			err:      nil,
			mockFlag: false,
			want: want{
				body:       "unauthorized",
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:      "Wrong user id type",
			taskID:    "task123",
			userIDCtx: 123,
			err:       nil,
			mockFlag:  false,
			want: want{
				body:       "userID has wrong type",
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.userIDCtx != nil {
					c.Set("userID", tc.userIDCtx)
				}
				c.Next()
			})

			r.DELETE("/tasks/:id", srv.deleteTask)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			repo := mocks.NewStorage(t)
			srv.db = repo

			taskDeleter := workers.NewTaskBatchDeleter(context.Background(), repo, 10, zerolog.Nop())
			srv.taskDeleter = taskDeleter

			if tc.mockFlag {
				repo.On(
					"MarkTaskToDelete",
					tc.taskID,
					"user123",
				).Return(tc.err)
			}

			req := resty.New().R()
			req.URL = httpSrv.URL + "/tasks/" + tc.taskID
			req.Method = http.MethodDelete

			res, err := req.Send()
			assert.NoError(t, err)
			assert.Equal(t, tc.want.statusCode, res.StatusCode())
			assert.Contains(t, string(res.Body()), tc.want.body)
		})
	}
}
