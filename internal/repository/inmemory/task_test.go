package inmemory

import (
	"testing"
	"toDoList/internal/domain/task/task_errors"
	"toDoList/internal/domain/task/task_models"

	"github.com/stretchr/testify/assert"
)

func TestStorage_Tasks(t *testing.T) {
	storage := NewInMemoryStorage()

	task1 := task_models.Task{
		ID: "task1",
		Attributes: task_models.TaskAttributes{
			Title:       "Task 1",
			Description: "First task",
			Status:      "New",
		},
		UserID: "user1",
	}

	tests := []struct {
		name        string
		action      func() error
		check       func(t *testing.T)
		expectError error
	}{
		{
			name: "AddTask_success",
			action: func() error {
				return storage.AddTask(task1)
			},
			check: func(t *testing.T) {
				assert.Len(t, storage.tasks, 1)
			},
			expectError: nil,
		},
		{
			name: "AddTask_duplicate",
			action: func() error {
				return storage.AddTask(task1)
			},
			check:       func(t *testing.T) {},
			expectError: task_errors.ErrorTaskIsAlreadyExist,
		},
		{
			name: "GetAllTasks_success",
			action: func() error {
				_, err := storage.GetAllTasks("user1")
				return err
			},
			check: func(t *testing.T) {
				tasks, _ := storage.GetAllTasks("user1")
				assert.Len(t, tasks, 1)
				assert.Equal(t, "task1", tasks[0].ID)
			},
			expectError: nil,
		},
		{
			name: "GetAllTasks_nothing_found",
			action: func() error {
				_, err := storage.GetAllTasks("unknown")
				return err
			},
			check:       func(t *testing.T) {},
			expectError: task_errors.FoundNothingErr,
		},
		{
			name: "GetTaskByID_success",
			action: func() error {
				_, err := storage.GetTaskByID("task1", "user1")
				return err
			},
			check: func(t *testing.T) {
				task, _ := storage.GetTaskByID("task1", "user1")
				assert.Equal(t, "Task 1", task.Attributes.Title)
			},
			expectError: nil,
		},
		{
			name: "GetTaskByID_not_found",
			action: func() error {
				_, err := storage.GetTaskByID("task404", "user1")
				return err
			},
			check:       func(t *testing.T) {},
			expectError: task_errors.FoundNothingErr,
		},
		{
			name: "UpdateTaskAttributes_success",
			action: func() error {
				task1.Attributes.Status = "Done"
				return storage.UpdateTaskAttributes(task1)
			},
			check: func(t *testing.T) {
				task, _ := storage.GetTaskByID("task1", "user1")
				assert.Equal(t, task_models.TaskStatus("Done"), task.Attributes.Status)
			},
			expectError: nil,
		},
		{
			name: "UpdateTaskAttributes_not_found",
			action: func() error {
				task := task_models.Task{ID: "unknown"}
				return storage.UpdateTaskAttributes(task)
			},
			check:       func(t *testing.T) {},
			expectError: task_errors.FoundNothingErr,
		},
		{
			name: "DeleteTask_success",
			action: func() error {
				return storage.DeleteTask("task1", "user1")
			},
			check: func(t *testing.T) {
				_, err := storage.GetTaskByID("task1", "user1")
				assert.ErrorIs(t, err, task_errors.FoundNothingErr)
			},
			expectError: nil,
		},
		{
			name: "DeleteTask_not_found",
			action: func() error {
				return storage.DeleteTask("task404", "user1")
			},
			check:       func(t *testing.T) {},
			expectError: task_errors.FoundNothingErr,
		},
		{
			name: "MarkTaskToDelete",
			action: func() error {
				return storage.MarkTaskToDelete("task2", "user2")
			},
			check:       func(t *testing.T) {},
			expectError: nil,
		},
		{
			name: "DeleteMarkedTasks",
			action: func() error {
				return storage.DeleteMarkedTasks()
			},
			check:       func(t *testing.T) {},
			expectError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.action()
			if tc.expectError != nil {
				assert.ErrorIs(t, err, tc.expectError)
			} else {
				assert.NoError(t, err)
			}
			tc.check(t)
		})
	}
}
