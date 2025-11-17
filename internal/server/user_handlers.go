package server

import (
	"net/http"
	"toDoList/internal/domain/user/user_errors"
	"toDoList/internal/domain/user/user_models"
	"toDoList/internal/service/user_service"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// TODO: + Админская ручка + форбидден для всех остальных
func (srv *ToDoListApi) getAllUsers(ctx *gin.Context) {
	usersService := user_service.NewUserService(srv.db)
	users, err := usersService.GetAllUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
	if len(users) != 0 {
		ctx.JSON(http.StatusOK, users)
	} else {
		ctx.JSON(http.StatusOK, "Task list is empty")
	}
}

// TODO: + Права для админа
func (srv *ToDoListApi) getUserByID(ctx *gin.Context) {
	userIDFromParam := ctx.Param("id")
	userIDFromCtx, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if userIDFromParam != userIDFromCtx.(string) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	usersService := user_service.NewUserService(srv.db)
	userInfo, err := usersService.GetUserByID(userIDFromParam)

	if err != nil {
		if errors.Is(err, user_errors.ErrorUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, userInfo)
}

func (srv *ToDoListApi) register(ctx *gin.Context) {
	var user user_models.UserRequest

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := user_service.NewUserService(srv.db)
	savedUser, err := service.SaveUser(user)
	if err != nil {
		if errors.Is(err, user_errors.ErrorUserIsAlreadyExist) {
			ctx.JSON(http.StatusConflict, gin.H{"error": user_errors.ErrorUserIsAlreadyExist.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	savedUser.Password = user.Password
	ctx.JSON(http.StatusOK, gin.H{"user": savedUser})
}

func (srv *ToDoListApi) login(ctx *gin.Context) {
	var usLogReq user_models.UserLoginRequest

	if err := ctx.ShouldBindJSON(&usLogReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := user_service.NewUserService(srv.db)
	user, err := service.LoginUser(usLogReq)
	if err != nil {
		if errors.Is(err, user_errors.ErrorInvalidPassword) || errors.Is(err, user_errors.ErrorUserNotExist) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": user_errors.ErrorNotValidCreds.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	access, err := srv.tokenSigner.NewAccessToken(user.Uuid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	refresh, err := srv.tokenSigner.NewRefreshToken(user.Uuid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.SetCookie("access_token", access, 3600*24, "/", "127.0.0.1:8080", false, true)
	ctx.SetCookie("refresh_token", refresh, 3600*24*7, "/", "127.0.0.1:8080", false, true)
	ctx.JSON(http.StatusOK, gin.H{"Message": "Login successful"})
}

func (srv *ToDoListApi) updateUser(ctx *gin.Context) {
	userIDFromParam := ctx.Param("id")
	userIDFromCtx, exists := ctx.Get("userID")

	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if userIDFromParam != userIDFromCtx.(string) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var newUser user_models.UserRequest
	if err := ctx.ShouldBindBodyWithJSON(&newUser); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	service := user_service.NewUserService(srv.db)
	newUserInfo, err := service.UpdateUser(userIDFromParam, newUser) // вернуть полного юзера
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"TaskID": newUserInfo})
}

func (srv *ToDoListApi) deleteUser(ctx *gin.Context) {
	userIDFromParam := ctx.Param("id")
	userIDFromCtx, exists := ctx.Get("userID")

	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if userIDFromParam != userIDFromCtx.(string) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	service := user_service.NewUserService(srv.db)
	if err := service.DeleteUser(userIDFromParam); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, "User was deleted")
}

func (srv *ToDoListApi) loginAdmin(ctx *gin.Context) {
	//TODO: получение всех тасок или всех юзеров - только под админскими правами
}
