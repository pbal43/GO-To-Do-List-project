package inmemory

import (
	"toDoList/internal/domain/task/taskerrors"
	"toDoList/internal/domain/task/taskmodels"
)

func (storage *Storage) GetAllTasks(userID string) ([]taskmodels.Task, error) {
	if len(storage.tasks) == 0 {
		return []taskmodels.Task{}, taskerrors.ErrFoundNothing
	}

	var tasks []taskmodels.Task

	for _, userTasks := range storage.tasks {
		if userTasks.UserID == userID {
			tasks = append(tasks, userTasks)
		}
	}

	if len(tasks) == 0 {
		return []taskmodels.Task{}, taskerrors.ErrFoundNothing
	}

	return tasks, nil
}
func (storage *Storage) GetTaskByID(taskID string, userID string) (taskmodels.Task, error) {
	if len(storage.tasks) == 0 {
		return taskmodels.Task{}, taskerrors.ErrFoundNothing
	}

	var task taskmodels.Task

	for _, userTasks := range storage.tasks {
		if userTasks.UserID == userID {
			if userTasks.ID == taskID {
				task = userTasks
				return task, nil
			}
		}
	}

	return taskmodels.Task{}, taskerrors.ErrFoundNothing
}

func (storage *Storage) AddTask(newTask taskmodels.Task) error {
	for _, t := range storage.tasks {
		if t.ID == newTask.ID {
			return taskerrors.ErrTaskIsAlreadyExist
		}
	}

	storage.tasks[newTask.ID] = newTask
	return nil
}

func (storage *Storage) UpdateTaskAttributes(task taskmodels.Task) error {
	for _, t := range storage.tasks {
		if t.ID == task.ID {
			t.Attributes = task.Attributes //nolint:govet // Нам не важно, копия это или нет
			storage.tasks[task.ID] = task
			return nil
		}
	}
	return taskerrors.ErrFoundNothing
}
func (storage *Storage) DeleteTask(taskID string, userID string) error {
	for _, t := range storage.tasks {
		if t.ID == taskID && t.UserID == userID {
			delete(storage.tasks, t.ID)
			return nil
		}
	}

	return taskerrors.ErrFoundNothing
}

//nolint:revive // Возможна будущая реализация, но нужно объявление для соблюдения интерфейса
func (storage *Storage) MarkTaskToDelete(taskID string, userID string) error {
	return nil
}

func (storage *Storage) DeleteMarkedTasks() error {
	return nil
}
