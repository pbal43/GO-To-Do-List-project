package inmemory

import (
	"toDoList/internal/domain/task/task_errors"
	"toDoList/internal/domain/task/task_models"
)

// TODO: протестить локальное хранилище

func (storage *Storage) GetAllTasks(userID string) ([]task_models.Task, error) {
	if len(storage.tasks) == 0 {
		return []task_models.Task{}, task_errors.FoundNothingErr
	}

	var tasks []task_models.Task

	for _, userTasks := range storage.tasks {
		if userTasks.UserID == userID {
			tasks = append(tasks, userTasks)
		}
	}

	if len(tasks) == 0 {
		return []task_models.Task{}, task_errors.FoundNothingErr
	}

	return tasks, nil
}
func (storage *Storage) GetTaskByID(taskID string, userID string) (task_models.Task, error) {
	if len(storage.tasks) == 0 {
		return task_models.Task{}, task_errors.FoundNothingErr
	}

	var task task_models.Task

	for _, userTasks := range storage.tasks {
		if userTasks.UserID == userID {
			if userTasks.ID == taskID {
				task = userTasks
				return task, nil
			}
		}
	}

	return task_models.Task{}, task_errors.FoundNothingErr
}

func (storage *Storage) AddTask(newTask task_models.Task) error {
	for _, t := range storage.tasks {
		if t.ID == newTask.ID {
			return task_errors.ErrorTaskIsAlreadyExist
		}
	}

	storage.tasks[newTask.ID] = newTask
	return nil
}

func (storage *Storage) UpdateTaskAttributes(task task_models.Task) error {
	for _, t := range storage.tasks {
		if t.ID == task.ID {
			t.Attributes = task.Attributes
			storage.tasks[task.ID] = task
			return nil
		}
	}
	return task_errors.FoundNothingErr
}
func (storage *Storage) DeleteTask(taskID string, userID string) error {
	for _, t := range storage.tasks {
		if t.ID == taskID && t.UserID == userID {
			delete(storage.tasks, t.ID)
			return nil
		}
	}

	return task_errors.FoundNothingErr
}

func (storage *Storage) MarkTaskToDelete(taskID string, userID string) error {
	return nil
}

func (storage *Storage) DeleteMarkedTasks() error {
	return nil
}
