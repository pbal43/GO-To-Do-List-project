package server

import (
	"net/http"
	"toDoList/internal"
	"toDoList/internal/domain/user/usererrors"
	"toDoList/internal/domain/user/usermodels"
	"toDoList/internal/service/userservice"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// TODO: + Админская ручка + форбидден для всех остальных
func (srv *ToDoListAPI) getAllUsers(ctx *gin.Context) {
	usersService := userservice.NewUserService(srv.db)
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
func (srv *ToDoListAPI) getUserByID(ctx *gin.Context) {
	userIDFromParam := ctx.Param("id")
	userIDFromCtx, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr, ok := userIDFromCtx.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error with userID"})
		return
	}

	if userIDFromParam != userIDStr {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	usersService := userservice.NewUserService(srv.db)
	userInfo, err := usersService.GetUserByID(userIDFromParam)

	if err != nil {
		if errors.Is(err, usererrors.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, userInfo)
}

func (srv *ToDoListAPI) register(ctx *gin.Context) {
	var user usermodels.UserRequest

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := userservice.NewUserService(srv.db)
	savedUser, err := service.SaveUser(user)
	if err != nil {
		if errors.Is(err, usererrors.ErrUserIsAlreadyExist) {
			ctx.JSON(http.StatusConflict, gin.H{"error": usererrors.ErrUserIsAlreadyExist.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	savedUser.Password = user.Password
	ctx.JSON(http.StatusOK, gin.H{"user": savedUser})
}

func (srv *ToDoListAPI) login(ctx *gin.Context) {
	var usLogReq usermodels.UserLoginRequest

	if err := ctx.ShouldBindJSON(&usLogReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := userservice.NewUserService(srv.db)
	user, err := service.LoginUser(usLogReq)
	if err != nil {
		if errors.Is(err, usererrors.ErrInvalidPassword) || errors.Is(err, usererrors.ErrUserNotExist) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": usererrors.ErrNotValidCreds.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	access, err := srv.tokenSigner.NewAccessToken(user.UUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	refresh, err := srv.tokenSigner.NewRefreshToken(user.UUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.SetCookie("access_token", access, internal.MaxAgeForAccessToken, "/", "127.0.0.1", false, true)
	ctx.SetCookie("refresh_token", refresh, internal.MaxAgeForRefreshToken, "/", "127.0.0.1", false, true)
	ctx.JSON(http.StatusOK, gin.H{"Message": "Login successful"})
}

func (srv *ToDoListAPI) updateUser(ctx *gin.Context) {
	userIDFromParam := ctx.Param("id")
	userIDFromCtx, exists := ctx.Get("userID")

	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr, ok := userIDFromCtx.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error with userID"})
		return
	}

	if userIDFromParam != userIDStr {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var newUser usermodels.UserRequest
	if err := ctx.ShouldBindBodyWithJSON(&newUser); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	service := userservice.NewUserService(srv.db)
	newUserInfo, err := service.UpdateUser(userIDFromParam, newUser)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"NewUserInfo": newUserInfo})
}

func (srv *ToDoListAPI) deleteUser(ctx *gin.Context) {
	userIDFromParam := ctx.Param("id")
	userIDFromCtx, exists := ctx.Get("userID")

	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr, ok := userIDFromCtx.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error with userID"})
		return
	}

	if userIDFromParam != userIDStr {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	service := userservice.NewUserService(srv.db)
	if err := service.DeleteUser(userIDFromParam); err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, "User was deleted")
}

//nolint:revive // Не реализовано
func (srv *ToDoListAPI) loginAdmin(ctx *gin.Context) {
	//TODO: получение всех тасок или всех юзеров - только под админскими правами
}
