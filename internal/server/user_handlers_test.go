package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"toDoList/internal/domain/user/user_errors"
	"toDoList/internal/domain/user/user_models"
	"toDoList/internal/server/mocks"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	var srv ToDoListApi
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.POST("/register", srv.register)
	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	type want struct {
		body       string
		statusCode int
	}

	type test struct {
		name       string
		userJson   string
		userFromDB user_models.User
		req        string
		method     string
		mockFlag   bool
		err        error
		want       want
	}

	tests := []test{
		{
			name:       "Register success",
			userJson:   `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: user_models.User{Uuid: "2246b7cc-4afa-4e31-abc9-24f8c95692f1", Name: "pere", Email: "pbsal@yaoo.com", Password: "unmarshall me"},
			req:        "/register",
			method:     http.MethodPost,
			mockFlag:   true,
			err:        nil,
			want: want{
				body:       `{"user":{"uuid":"2246b7cc-4afa-4e31-abc9-24f8c95692f1","name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}}`,
				statusCode: http.StatusOK,
			},
		},
		{
			name:     "Bad JSON",
			userJson: `{name=jsonDoesNotExist}`,
			req:      "/register",
			method:   http.MethodPost,
			mockFlag: false,
			err:      fmt.Errorf(`invalid character 'n' looking for beginning of value`),
			want: want{
				statusCode: http.StatusBadRequest,
				body:       `{"error":"invalid character`,
			},
		},
		{
			name:       "User already exists",
			userJson:   `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: user_models.User{Uuid: "2246b7cc-4afa-4e31-abc9-24f8c95692f1", Name: "pere", Email: "pbsal@yaoo.com", Password: "unmarshall me"},
			req:        "/register",
			method:     http.MethodPost,
			mockFlag:   true,
			err:        user_errors.ErrorUserIsAlreadyExist,
			want: want{
				body:       "user is already exist",
				statusCode: http.StatusConflict,
			},
		},
		{
			name:       "Internal server error",
			userJson:   `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: user_models.User{Uuid: "2246b7cc-4afa-4e31-abc9-24f8c95692f1", Name: "pere", Email: "pbsal@yaoo.com", Password: "unmarshall me"},
			req:        "/register",
			method:     http.MethodPost,
			mockFlag:   true,
			err:        fmt.Errorf(`internal server error`),
			want: want{
				body:       "internal server error",
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			srv.db = repo
			if tc.mockFlag {
				repo.On("SaveUser", mock.MatchedBy(func(user user_models.User) bool {
					return user.Name == tc.userFromDB.Name && user.Email == tc.userFromDB.Email
				})).Return(tc.userFromDB, tc.err)

			}
			req := resty.New().R()
			req.URL = httpSrv.URL + tc.req
			req.Method = tc.method
			req.Body = tc.userJson

			res, err := req.Send()
			assert.NoError(t, err)
			assert.Equal(t, tc.want.statusCode, res.StatusCode())
			if tc.err == nil {
				assert.Equal(t, tc.want.body, string(res.Body()))
			} else {
				assert.Contains(t, string(res.Body()), tc.want.body)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	var srv ToDoListApi
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.POST("/login", srv.login)
	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	testPass := "unmarshall me"
	testPassHash, err := bcrypt.GenerateFromPassword([]byte(testPass), bcrypt.DefaultCost)
	assert.NoError(t, err)

	type want struct {
		cookie     bool
		body       string
		statusCode int
	}

	type test struct {
		name        string
		userJson    string
		userRequest user_models.UserLoginRequest
		userFromDB  user_models.User
		req         string
		method      string
		mockFlag    bool
		want        want
	}

	tests := []test{
		{
			name:        "Login success",
			userJson:    fmt.Sprintf(`{"email":"pbsal@yaoo.com","password":"%s"}`, testPass),
			userRequest: user_models.UserLoginRequest{Email: "pbsal@yaoo.com", Password: "unmarshall me"},
			userFromDB:  user_models.User{Uuid: "2246b7cc-4afa-4e31-abc9-24f8c95692f1", Name: "pere", Email: "pbsal@yaoo.com", Password: string(testPassHash)},
			req:         "/login",
			method:      http.MethodPost,
			mockFlag:    true,
			want: want{
				cookie:     true,
				body:       `{"Message":"Login successful"}`,
				statusCode: http.StatusOK,
			},
		},
		{
			name:        "Unauthorized",
			userJson:    `{"email":"pbsal@yaoo.com","password":"wrongPassword"}`,
			userRequest: user_models.UserLoginRequest{Email: "pbsal@yaoo.com", Password: "unmarshall me"},
			userFromDB:  user_models.User{Uuid: "2246b7cc-4afa-4e31-abc9-24f8c95692f1", Name: "pere", Email: "pbsal@yaoo.com", Password: string(testPassHash)},
			req:         "/login",
			method:      http.MethodPost,
			mockFlag:    true,
			want: want{
				cookie:     true,
				body:       `{"error":"the creds are invalid"}`,
				statusCode: http.StatusUnauthorized,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			srv.db = repo
			if tc.mockFlag {
				repo.On("GetUserByEmail", tc.userRequest.Email).Return(tc.userFromDB, nil)
			}
			req := resty.New().R()
			req.URL = httpSrv.URL + tc.req
			req.Method = tc.method
			req.Body = tc.userJson

			res, err := req.Send()

			assert.NoError(t, err)
			assert.Equal(t, tc.want.statusCode, res.StatusCode())
			assert.Equal(t, tc.want.body, string(res.Body()))
			assert.NotNil(t, res.Cookies())
		})
	}
}
