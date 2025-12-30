package taskservice

import (
	"toDoList/internal/domain/task/taskerrors"
	"toDoList/internal/domain/task/taskmodels"
	"toDoList/internal/server/workers"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TaskStorage interface {
	GetAllTasks(userID string) ([]taskmodels.Task, error)
	GetTaskByID(taskID string, userID string) (taskmodels.Task, error)
	AddTask(newTask taskmodels.Task) error
	UpdateTaskAttributes(task taskmodels.Task) error
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

func (ts *TaskService) GetAllTasks(userID string) ([]taskmodels.Task, error) {
	return ts.db.GetAllTasks(userID)
}

func (ts *TaskService) GetTaskByID(taskID string, userID string) (taskmodels.Task, error) {
	if taskID == "" {
		return taskmodels.Task{}, taskerrors.ErrEmptyString
	}

	task, err := ts.db.GetTaskByID(taskID, userID)
	if err != nil {
		return taskmodels.Task{}, err
	}

	return task, nil
}

func (ts *TaskService) CreateTask(newTaskAttributes taskmodels.TaskAttributes, userID string) (string, error) {
	err := ts.valid.Struct(newTaskAttributes)
	if err != nil {
		return "", err
	}

	taskStatusValid := newTaskAttributes.Status.IsValid()

	if !taskStatusValid {
		return "", taskerrors.ErrWrongStatus
	}

	var newTask taskmodels.Task

	newTask.ID = uuid.New().String()
	newTask.UserID = userID
	newTask.Attributes = newTaskAttributes

	err = ts.db.AddTask(newTask)
	if err != nil {
		return "", err
	}

	return newTask.ID, nil
}

func (ts *TaskService) UpdateTask(taskID string, userID string, newAttributes taskmodels.TaskAttributes) error {
	err := ts.valid.Struct(newAttributes)
	if err != nil {
		return err
	}

	taskStatusValid := newAttributes.Status.IsValid()

	if !taskStatusValid {
		return taskerrors.ErrWrongStatus
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
