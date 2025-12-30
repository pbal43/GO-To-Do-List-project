package server

import (
	"fmt"
	"net/http"
	"toDoList/internal/domain/task/taskmodels"
	"toDoList/internal/service/taskservice"

	"github.com/gin-gonic/gin"
)

// обрабатываем для вывода, возвращаем респонсы с ошибками и проч.

func (srv *ToDoListAPI) getTasks(ctx *gin.Context) {
	userIDFromCtx, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDFromCtx.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "userID has wrong type"})
		return
	}

	taskService := taskservice.NewTaskService(srv.db, srv.taskDeleter)
	tasks, err := taskService.GetAllTasks(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
	if len(tasks) != 0 {
		ctx.JSON(http.StatusOK, tasks)
	} else {
		ctx.JSON(http.StatusOK, "Task list is empty")
	}
}

func (srv *ToDoListAPI) getTaskByID(ctx *gin.Context) {
	taskID := ctx.Param("id")

	userIDFromCtx, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDFromCtx.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "userID has wrong type"})
		return
	}

	taskService := taskservice.NewTaskService(srv.db, srv.taskDeleter)
	foundedTask, err := taskService.GetTaskByID(taskID, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, foundedTask)
}

func (srv *ToDoListAPI) createTask(ctx *gin.Context) {
	userIDFromCtx, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDFromCtx.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "userID has wrong type"})
		return
	}

	var newTaskAttributes taskmodels.TaskAttributes
	if err := ctx.ShouldBindBodyWithJSON(&newTaskAttributes); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	taskService := taskservice.NewTaskService(srv.db, srv.taskDeleter)
	taskID, err := taskService.CreateTask(newTaskAttributes, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"TaskID": taskID})
}

func (srv *ToDoListAPI) updateTask(ctx *gin.Context) {
	taskID := ctx.Param("id")

	userIDFromCtx, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDFromCtx.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "userID has wrong type"})
		return
	}

	var newAttributes taskmodels.TaskAttributes
	if err := ctx.ShouldBindBodyWithJSON(&newAttributes); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	taskService := taskservice.NewTaskService(srv.db, srv.taskDeleter)
	err := taskService.UpdateTask(taskID, userID, newAttributes)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, fmt.Sprintf("TaskID: %s was updated", taskID))
}

func (srv *ToDoListAPI) deleteTask(ctx *gin.Context) {
	taskID := ctx.Param("id")
	userIDFromCtx, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDFromCtx.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "userID has wrong type"})
		return
	}

	taskService := taskservice.NewTaskService(srv.db, srv.taskDeleter)
	if err := taskService.MarkTaskToDeleteByID(taskID, userID); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, "Task was deleted")
}
