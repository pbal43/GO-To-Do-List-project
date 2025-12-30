package userservice

import (
	"testing"
	"toDoList/internal/domain/user/usererrors"
	"toDoList/internal/domain/user/usermodels"
	"toDoList/internal/server/mocks"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestGetAllUsers(t *testing.T) {
	type want struct {
		usersData []usermodels.User
		err       error
	}

	type test struct {
		name        string
		dataFromDB  []usermodels.User
		errorFromDB error
		want        want
	}

	tests := []test{
		{
			name: "success",
			dataFromDB: []usermodels.User{
				{
					UUID:     "1",
					Name:     "John Doe",
					Email:    "test",
					Password: "password",
				},
				{
					UUID:     "2",
					Name:     "John Doe",
					Email:    "test",
					Password: "password",
				},
			},
			errorFromDB: nil,
			want: want{
				usersData: []usermodels.User{{
					UUID:     "1",
					Name:     "John Doe",
					Email:    "test",
					Password: "password",
				},
					{
						UUID:     "2",
						Name:     "John Doe",
						Email:    "test",
						Password: "password",
					},
				},
				err: nil,
			},
		},
		{
			name:        "error",
			dataFromDB:  []usermodels.User{},
			errorFromDB: usererrors.ErrGetAllUsersData,
			want: want{
				usersData: []usermodels.User{},
				err:       usererrors.ErrGetAllUsersData,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			newService := NewUserService(repo)

			repo.On("GetAllUsers").Return(tc.dataFromDB, tc.errorFromDB)
			users, err := newService.GetAllUsers()

			assert.Equal(t, tc.want.usersData, users)
			assert.Equal(t, tc.want.err, err)
		})
	}
}

func TestGetUserById(t *testing.T) {
	type want struct {
		usersData usermodels.User
		err       error
	}

	type test struct {
		name        string
		userID      string
		dataFromDB  usermodels.User
		errorFromDB error
		dbMock      bool
		want        want
	}

	tests := []test{
		{
			name:   "success",
			userID: "1",
			dataFromDB: usermodels.User{
				UUID:     "1",
				Name:     "John Doe",
				Email:    "test",
				Password: "password",
			},
			errorFromDB: nil,
			dbMock:      true,
			want: want{
				usersData: usermodels.User{
					UUID:     "1",
					Name:     "John Doe",
					Email:    "test",
					Password: "password",
				},
				err: nil,
			},
		},
		{
			name:   "fail: userID is empty string",
			userID: "",
			dbMock: false,
			want: want{
				usersData: usermodels.User{},
				err:       usererrors.ErrUserEmptyInsert,
			},
		},
		{
			name:        "err from database",
			userID:      "1",
			dataFromDB:  usermodels.User{},
			errorFromDB: usererrors.ErrUserNotExist,
			dbMock:      true,
			want: want{
				usersData: usermodels.User{},
				err:       usererrors.ErrUserNotExist,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			newService := NewUserService(repo)

			if tc.dbMock {
				repo.On("GetUserByID", tc.userID).Return(tc.dataFromDB, tc.errorFromDB)
			}

			users, err := newService.GetUserByID(tc.userID)

			assert.Equal(t, tc.want.usersData, users)
			assert.Equal(t, tc.want.err, err)
		})
	}
}

func TestSaveUser(t *testing.T) {
	type want struct {
		usersData usermodels.User
		err       error
	}

	type test struct {
		name        string
		UserRequest usermodels.UserRequest
		dataFromDB  usermodels.User
		errorFromDB error
		validErr    bool
		dbMock      bool
		want        want
	}

	tests := []test{
		{
			name: "success",
			UserRequest: usermodels.UserRequest{
				Name:     "John Doe",
				Email:    "test@test.ru",
				Password: "password123!",
			},
			dataFromDB: usermodels.User{
				UUID:     "1",
				Name:     "John Doe",
				Email:    "test",
				Password: "password",
			},
			errorFromDB: nil,
			dbMock:      true,
			want: want{
				usersData: usermodels.User{
					Name:     "John Doe",
					Email:    "test",
					Password: "password",
				},
				err: nil,
			},
		},
		{
			name: "error not valid user request data",
			UserRequest: usermodels.UserRequest{
				Name:     "John Doe",
				Email:    "test",
				Password: "password123!",
			},
			dbMock:   false,
			validErr: true,
			want: want{
				usersData: usermodels.User{},
				err:       nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			newService := NewUserService(repo)

			if tc.dbMock {
				repo.On("SaveUser", mock.Anything).Return(tc.dataFromDB, tc.errorFromDB)
			}

			users, err := newService.SaveUser(tc.UserRequest)

			assert.Equal(t, tc.want.usersData.Name, users.Name)
			assert.Equal(t, tc.want.usersData.Email, users.Email)
			assert.Equal(t, tc.want.usersData.Password, users.Password)

			if tc.validErr {
				assert.IsType(t, err, validator.ValidationErrors{})
			} else {
				assert.Equal(t, tc.want.err, err)
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	type want struct {
		usersData usermodels.User
		err       error
	}

	type test struct {
		name             string
		UserLoginRequest usermodels.UserLoginRequest
		dataFromDB       usermodels.User
		errorFromDB      error
		dbMock           bool
		needPwdHash      bool
		needPwdClean     bool
		want             want
	}

	tests := []test{
		{
			name: "success",
			UserLoginRequest: usermodels.UserLoginRequest{
				Email:    "test@test.ru",
				Password: "password123!",
			},
			dataFromDB: usermodels.User{
				UUID:     "1",
				Name:     "John Doe",
				Email:    "test@test.ru",
				Password: "password123!",
			},
			errorFromDB: nil,
			dbMock:      true,
			needPwdHash: true,
			want: want{
				usersData: usermodels.User{
					UUID:     "1",
					Name:     "John Doe",
					Email:    "test@test.ru",
					Password: "password123!",
				},
				err: nil,
			},
		},
		{
			name: "error from db",
			UserLoginRequest: usermodels.UserLoginRequest{
				Email:    "test@test.ru",
				Password: "password123!",
			},
			dataFromDB:  usermodels.User{},
			errorFromDB: usererrors.ErrUserNotExist,
			dbMock:      true,
			needPwdHash: false,
			want: want{
				usersData: usermodels.User{},
				err:       usererrors.ErrUserNotExist,
			},
		},
		{
			name: "fail: wrong password",
			UserLoginRequest: usermodels.UserLoginRequest{
				Email:    "test@test.ru",
				Password: "wrongPassword123!",
			},
			dataFromDB: usermodels.User{
				UUID:     "1",
				Name:     "John Doe",
				Email:    "test@test.ru",
				Password: "password123!",
			},
			errorFromDB:  nil,
			dbMock:       true,
			needPwdHash:  true,
			needPwdClean: true,
			want: want{
				usersData: usermodels.User{},
				err:       usererrors.ErrInvalidPassword,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			newService := NewUserService(repo)

			if tc.needPwdHash {
				hash, _ := bcrypt.GenerateFromPassword([]byte(tc.dataFromDB.Password), bcrypt.DefaultCost)
				tc.dataFromDB.Password = string(hash)
				tc.want.usersData.Password = string(hash)
			}

			if tc.needPwdClean {
				tc.want.usersData.Password = ""
			}

			if tc.dbMock {
				repo.On("GetUserByEmail", tc.UserLoginRequest.Email).Return(tc.dataFromDB, tc.errorFromDB)
			}

			users, err := newService.LoginUser(tc.UserLoginRequest)

			assert.Equal(t, tc.want.usersData, users)
			assert.Equal(t, tc.want.err, err)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	type want struct {
		usersData usermodels.User
		err       error
	}

	type test struct {
		name              string
		userID            string
		userRequest       usermodels.UserRequest
		dataFromDBGet     usermodels.User
		errorFromDBGet    error
		dataFromDBUpdate  usermodels.User
		errorFromDBUpdate error
		dbMockGet         bool
		dbMockUpdate      bool
		validErr          bool
		want              want
	}

	tests := []test{
		{
			name:   "success",
			userID: "lalala",
			userRequest: usermodels.UserRequest{
				Name:     "Petro",
				Email:    "petr@petr.ru",
				Password: "password123!",
			},
			dataFromDBGet: usermodels.User{
				UUID:     "lalala",
				Name:     "John Doe",
				Email:    "john@john.com",
				Password: "oldpassword123!",
			},
			errorFromDBGet: nil,
			dataFromDBUpdate: usermodels.User{
				UUID:     "lalala",
				Name:     "John Doe",
				Email:    "petr@petr.ru",
				Password: "password123!",
			},
			errorFromDBUpdate: nil,
			dbMockGet:         true,
			dbMockUpdate:      true,
			want: want{
				usersData: usermodels.User{
					UUID:     "lalala",
					Name:     "John Doe",
					Email:    "petr@petr.ru",
					Password: "password123!",
				},
				err: nil,
			},
		},
		{
			name:   "fail: validation error",
			userID: "lalala",
			userRequest: usermodels.UserRequest{
				Name:     "Petro",
				Email:    "petr@petr.ru",
				Password: "passw",
			},
			errorFromDBUpdate: nil,
			dbMockGet:         false,
			dbMockUpdate:      false,
			validErr:          true,
			want: want{
				usersData: usermodels.User{},
				err:       nil,
			},
		},
		{
			name:   "fail: get db error",
			userID: "lalala",
			userRequest: usermodels.UserRequest{
				Name:     "Petro",
				Email:    "petr@petr.ru",
				Password: "password123!",
			},
			dataFromDBGet: usermodels.User{
				UUID:     "lalala",
				Name:     "John Doe",
				Email:    "john@john.com",
				Password: "oldpassword123!",
			},
			errorFromDBGet: usererrors.ErrUserNotExist,
			dbMockGet:      true,
			want: want{
				usersData: usermodels.User{},
				err:       usererrors.ErrUserNotExist,
			},
		},
		{
			name:   "fail: update db error",
			userID: "lalala",
			userRequest: usermodels.UserRequest{
				Name:     "Petro",
				Email:    "petr@petr.ru",
				Password: "password123!",
			},
			dataFromDBGet: usermodels.User{
				UUID:     "lalala",
				Name:     "John Doe",
				Email:    "john@john.com",
				Password: "oldpassword123!",
			},
			errorFromDBGet:    nil,
			dataFromDBUpdate:  usermodels.User{},
			errorFromDBUpdate: usererrors.ErrUserNotFound,
			dbMockGet:         true,
			dbMockUpdate:      true,
			want: want{
				usersData: usermodels.User{},
				err:       usererrors.ErrUserNotFound,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			newService := NewUserService(repo)

			if tc.dbMockGet {
				repo.On("GetUserByID", tc.userID).Return(tc.dataFromDBGet, tc.errorFromDBGet)
			}

			if tc.dbMockUpdate {
				repo.On("UpdateUser", mock.Anything).Return(tc.dataFromDBUpdate, tc.errorFromDBUpdate)
			}

			users, err := newService.UpdateUser(tc.userID, tc.userRequest)

			assert.Equal(t, tc.want.usersData, users)

			if tc.validErr {
				assert.IsType(t, err, validator.ValidationErrors{})
			} else {
				assert.Equal(t, tc.want.err, err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	type want struct {
		err error
	}

	type test struct {
		name        string
		userID      string
		errorFromDB error
		want        want
	}

	tests := []test{
		{
			name:        "success",
			userID:      "lalala",
			errorFromDB: nil,
			want: want{
				err: nil,
			},
		},
		{
			name:        "fail: db error",
			userID:      "lalala",
			errorFromDB: usererrors.ErrUserNotFound,
			want: want{
				err: usererrors.ErrUserNotFound,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			newService := NewUserService(repo)

			repo.On("DeleteUser", tc.userID).Return(tc.errorFromDB)

			err := newService.DeleteUser(tc.userID)

			assert.Equal(t, tc.want.err, err)
		})
	}
}
