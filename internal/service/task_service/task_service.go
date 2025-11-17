package task_service

import (
	"toDoList/internal/domain/task/task_errors"
	"toDoList/internal/domain/task/task_models"
	"toDoList/internal/server/workers"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TaskStorage interface {
	GetAllTasks(userID string) ([]task_models.Task, error)
	GetTaskByID(taskID string, userID string) (task_models.Task, error)
	AddTask(newTask task_models.Task) error
	UpdateTaskAttributes(task task_models.Task) error
	DeleteTask(taskID string, userID string) error
	MarkTaskToDelete(taskID string, userID string) error
}

type TaskService struct {
	db          TaskStorage
	valid       *validator.Validate
	taskDeleter *workers.TaskBatchDeleter
}

func NewTaskService(db TaskStorage, taskDeleter *workers.TaskBatchDeleter) *TaskService {
	return &TaskService{db: db, valid: validator.New(), taskDeleter: taskDeleter}
}

func (ts *TaskService) GetAllTasks(userID string) ([]task_models.Task, error) {
	return ts.db.GetAllTasks(userID)
}

func (ts *TaskService) GetTaskByID(taskID string, userID string) (task_models.Task, error) {
	if taskID == "" {
		return task_models.Task{}, task_errors.EpmtyStringErr
	}

	task, err := ts.db.GetTaskByID(taskID, userID)
	if err != nil {
		return task_models.Task{}, err
	}

	return task, nil
}

func (ts *TaskService) CreateTask(newTaskAttributes task_models.TaskAttributes, userID string) (string, error) {
	err := ts.valid.Struct(newTaskAttributes)
	if err != nil {
		return "", err
	}

	taskStatusValid := newTaskAttributes.Status.IsValid()

	if !taskStatusValid {
		return "", task_errors.WrongStatusErr
	}

	var newTask task_models.Task

	newTask.ID = uuid.New().String()
	newTask.UserID = userID
	newTask.Attributes = newTaskAttributes

	err = ts.db.AddTask(newTask)
	if err != nil {
		return "", err
	}

	return newTask.ID, nil
}

func (ts *TaskService) UpdateTask(taskID string, userID string, newAttributes task_models.TaskAttributes) error {
	err := ts.valid.Struct(newAttributes)
	if err != nil {
		return err
	}

	taskStatusValid := newAttributes.Status.IsValid()

	if !taskStatusValid {
		return task_errors.WrongStatusErr
	}

	task, err := ts.db.GetTaskByID(taskID, userID)
	if err != nil {
		return err
	}

	task.Attributes = newAttributes

	err = ts.db.UpdateTaskAttributes(task)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TaskService) DeleteTaskByID(taskID string, userID string) error {
	err := ts.db.DeleteTask(taskID, userID)
	if err != nil {
		return err
	}
	return nil
}

func (ts *TaskService) MarkTaskToDeleteByID(taskID string, userID string) error {
	err := ts.db.MarkTaskToDelete(taskID, userID)
	if err != nil {
		return err
	}
	ts.taskDeleter.Notify()
	return nil
}
