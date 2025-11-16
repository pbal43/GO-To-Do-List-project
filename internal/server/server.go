package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"toDoList/internal"
	"toDoList/internal/domain/task/task_models"
	"toDoList/internal/domain/user/user_models"
	auth "toDoList/internal/server/auth/user_auth"
	"toDoList/internal/server/middleware"
	"toDoList/internal/server/workers"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
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

type ToDoListApi struct {
	srv         *http.Server
	db          Storage
	tokenSigner auth.HS256Signer
	taskDeleter *workers.TaskBatchDeleter
}

func NewServer(cfg internal.Config, db Storage, taskDeleter *workers.TaskBatchDeleter) *ToDoListApi {

	signer := auth.HS256Signer{
		Secret:     []byte("ultraSecretKey123"),
		Issuer:     "todolistService",
		Audience:   "todolistClient",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 24 * 7 * time.Hour,
	}

	HttpSrv := http.Server{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}

	api := ToDoListApi{srv: &HttpSrv, db: db, tokenSigner: signer, taskDeleter: taskDeleter}

	api.configRouter()

	return &api
}

func (api *ToDoListApi) Run() error {
	return api.srv.ListenAndServe()
}

func (api *ToDoListApi) ShutDown(ctx context.Context) error {
	return api.srv.Shutdown(ctx)
}

func (api *ToDoListApi) configRouter() {
	router := gin.Default()

	router.Use(middleware.GzipDecompressMiddleware())

	router.Use(gzip.Gzip(gzip.DefaultCompression,
		gzip.WithExcludedExtensions([]string{".png", ".jpg", ".gif", ".mp4"}),
	))

	tasks := router.Group("/tasks")
	{
		tasks.GET("/", middleware.AuthMiddleware(api.tokenSigner), api.getTasks)
		tasks.GET("/:id", middleware.AuthMiddleware(api.tokenSigner), api.getTaskByID)
		tasks.POST("/", middleware.AuthMiddleware(api.tokenSigner), api.createTask)
		tasks.PUT("/:id", middleware.AuthMiddleware(api.tokenSigner), api.updateTask)
		tasks.DELETE("/:id", middleware.AuthMiddleware(api.tokenSigner), api.deleteTask)
	}

	users := router.Group("/users")
	{
		users.GET("/", api.getAllUsers) // TODO: чисто админская
		users.GET("/:id", middleware.AuthMiddleware(api.tokenSigner), api.getUserByID)
		users.POST("/register", api.register)
		users.POST("/login", api.login)
		users.POST("/admin-login", api.loginAdmin)
		users.PUT("/:id", middleware.AuthMiddleware(api.tokenSigner), api.updateUser)
		users.DELETE("/:id", middleware.AuthMiddleware(api.tokenSigner), api.deleteUser)
	}

	api.srv.Handler = router
}
