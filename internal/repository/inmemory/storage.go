package inmemory

import (
	"toDoList/internal/domain/task/taskmodels"
	"toDoList/internal/domain/user/usermodels"
)

type Storage struct {
	users map[string]usermodels.User
	tasks map[string]taskmodels.Task
}

func NewInMemoryStorage() *Storage {
	return &Storage{
		users: make(map[string]usermodels.User),
		tasks: make(map[string]taskmodels.Task),
	}
}
