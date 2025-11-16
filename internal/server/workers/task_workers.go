package workers

import (
	"context"
	"time"
	"toDoList/internal/domain/task/task_models"
	"toDoList/internal/domain/user/user_models"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type UserStorage interface {
	GetAllUsers() ([]user_models.User, error)
	SaveUser(user user_models.User) (user_models.User, error)
	GetUserByID(userID string) (user_models.User, error)
	GetUserByEmail(email string) (user_models.User, error)
	UpdateUser(user user_models.User) (user_models.User, error)
	DeleteUser(userID string) error
}

type TaskStorage interface {
	GetAllTasks(userID string) ([]task_models.Task, error)
	GetTaskByID(taskID string, userID string) (task_models.Task, error)
	AddTask(newTask task_models.Task) error
	UpdateTaskAttributes(task task_models.Task) error
	DeleteTask(taskID string, userID string) error
	MarkTaskToDelete(taskID string, userID string) error
	DeleteMarkedTasks() error
}

type Storage interface {
	UserStorage
	TaskStorage
}

type TaskBatchDeleter struct {
	storage  Storage
	taskChan chan struct{}
	capacity int
	ctx      context.Context
	log      zerolog.Logger
}

func NewTaskBatchDeleter(storage Storage, ctx context.Context, capacity int, log zerolog.Logger) *TaskBatchDeleter {
	return &TaskBatchDeleter{
		storage:  storage,
		taskChan: make(chan struct{}, capacity),
		capacity: capacity,
		ctx:      ctx,
		log:      log,
	}
}

func (t *TaskBatchDeleter) Start() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			t.log.Info().Msg("TaskBatchDeleter stopped")
			return
		case <-ticker.C:
			if len(t.taskChan) == t.capacity {
				err := t.deleteTasks()
				if err != nil {
					log.Error().Err(err).Msg("failed to delete tasks")
				}
			}
		}
	}
}

func (t *TaskBatchDeleter) Stop() error {
	err := t.deleteTasks()
	if err != nil {
		return err
	}
	close(t.taskChan)
	return nil
}

func (t *TaskBatchDeleter) deleteTasks() error {
	t.flushChannel()
	err := t.storage.DeleteMarkedTasks()
	if err != nil {
		return err
	}
	return nil
}

func (t *TaskBatchDeleter) Notify() {
	select {
	case t.taskChan <- struct{}{}:
		t.log.Info().Msg("add task to delete queue")
	default:
		t.log.Warn().Msg("taskChan full, skipping notify") // мягко относимся к переполнению, не так важно
	}
}

func (t *TaskBatchDeleter) flushChannel() {
	for {
		select {
		case <-t.taskChan:
			t.log.Info().Msg("flushing tasks")
		default:
			return
		}
	}
}
