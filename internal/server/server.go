package server

import (
	"context"
	"fmt"
	"net/http"
	"toDoList/internal"
	"toDoList/internal/domain/task/taskmodels"
	"toDoList/internal/domain/user/usermodels"
	auth "toDoList/internal/server/auth/user_auth"
	"toDoList/internal/server/middleware"
	"toDoList/internal/server/workers"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type UserStorage interface {
	GetAllUsers() ([]usermodels.User, error)
	SaveUser(user usermodels.User) (usermodels.User, error)
	GetUserByID(userID string) (usermodels.User, error)
	GetUserByEmail(email string) (usermodels.User, error)
	UpdateUser(user usermodels.User) (usermodels.User, error)
	DeleteUser(userID string) error
}

type TaskStorage interface {
	GetAllTasks(userID string) ([]taskmodels.Task, error)
	GetTaskByID(taskID string, userID string) (taskmodels.Task, error)
	AddTask(newTask taskmodels.Task) error
	UpdateTaskAttributes(task taskmodels.Task) error
	DeleteTask(taskID string, userID string) error
	MarkTaskToDelete(taskID string, userID string) error
	DeleteMarkedTasks() error
}

type Storage interface {
	UserStorage
	TaskStorage
}

type TokenSigner interface {
	NewAccessToken(userID string) (string, error)
	NewRefreshToken(userID string) (string, error)
	ParseAccessToken(token string, opt auth.ParseOptions) (*auth.Claims, error)
	ParseRefreshToken(token string, opt auth.ParseOptions) (*jwt.RegisteredClaims, error)
	GetIssuer() string
	GetAudience() string
}

type ToDoListAPI struct {
	srv         *http.Server
	db          Storage
	tokenSigner TokenSigner
	taskDeleter *workers.TaskBatchDeleter
}

func NewServer(
	cfg internal.Config,
	db Storage,
	tokenSigner TokenSigner,
	taskDeleter *workers.TaskBatchDeleter,
) *ToDoListAPI {
	HTTPSrv := http.Server{ //nolint:gocritic // Линтеры противоречат друг другу, оставил так
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		ReadHeaderTimeout: internal.SecFive,
	}

	api := ToDoListAPI{srv: &HTTPSrv, db: db, tokenSigner: tokenSigner, taskDeleter: taskDeleter}

	api.configRouter()

	return &api
}

func (api *ToDoListAPI) Run() error {
	return api.srv.ListenAndServe()
}

func (api *ToDoListAPI) ShutDown(ctx context.Context) error {
	return api.srv.Shutdown(ctx)
}

func (api *ToDoListAPI) configRouter() {
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
