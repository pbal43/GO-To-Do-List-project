package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"toDoList/internal/domain/user/usererrors"
	"toDoList/internal/domain/user/usermodels"
	"toDoList/internal/server/mocks"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	var srv ToDoListAPI
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
		userJSON   string
		userFromDB usermodels.User
		req        string
		method     string
		mockFlag   bool
		err        error
		want       want
	}

	tests := []test{
		{
			name:     "Register success",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:      "/register",
			method:   http.MethodPost,
			mockFlag: true,
			err:      nil,
			want: want{
				body:       `{"user":{"uuid":"2246b7cc-4afa-4e31-abc9-24f8c95692f1","name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}}`,
				statusCode: http.StatusOK,
			},
		},
		{
			name:     "Bad JSON",
			userJSON: `{name=jsonDoesNotExist}`,
			req:      "/register",
			method:   http.MethodPost,
			mockFlag: false,
			err:      usererrors.ErrInvalidChar,
			want: want{
				statusCode: http.StatusBadRequest,
				body:       `{"error":"invalid character`,
			},
		},
		{
			name:     "User already exists",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:      "/register",
			method:   http.MethodPost,
			mockFlag: true,
			err:      usererrors.ErrUserIsAlreadyExist,
			want: want{
				body:       "user is already exist",
				statusCode: http.StatusConflict,
			},
		},
		{
			name:     "Internal server error",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:      "/register",
			method:   http.MethodPost,
			mockFlag: true,
			err:      usererrors.ErrInternalServer,
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
				repo.On("SaveUser", mock.MatchedBy(func(user usermodels.User) bool {
					return user.Name == tc.userFromDB.Name && user.Email == tc.userFromDB.Email
				})).Return(tc.userFromDB, tc.err)
			}
			req := resty.New().R()
			req.URL = httpSrv.URL + tc.req
			req.Method = tc.method
			req.Body = tc.userJSON

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
	var srv ToDoListAPI
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.POST("/login", srv.login)
	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	testPass := "unmarshall me"
	testPassHash, err := bcrypt.GenerateFromPassword([]byte(testPass), bcrypt.DefaultCost)
	assert.NoError(t, err)

	testToken := "any string"

	type want struct {
		cookie     bool
		body       string
		statusCode int
	}

	type TokenSignerResp struct {
		accessToken       string
		accessTokenError  error
		refreshToken      string
		refreshTokenError error
	}

	type test struct {
		name                       string
		userJSON                   string
		userRequest                usermodels.UserLoginRequest
		userFromDB                 usermodels.User
		req                        string
		method                     string
		mockFlagDB                 bool
		mockFlagTokenSignerAccess  bool
		mockFlagTokenSignerRefresh bool
		TokenSignerResponse        TokenSignerResp
		TokenSignerAccessResp      string

		err  error
		want want
	}

	tests := []test{
		{
			name: "Login success",
			userJSON: fmt.Sprintf(
				`{"email":"pbsal@yaoo.com","password":"%s"}`,
				testPass,
			),
			userRequest: usermodels.UserLoginRequest{
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: string(testPassHash),
			},
			req:                        "/login",
			method:                     http.MethodPost,
			mockFlagDB:                 true,
			mockFlagTokenSignerAccess:  true,
			mockFlagTokenSignerRefresh: true,
			err:                        nil,
			TokenSignerResponse: TokenSignerResp{
				accessToken:  testToken,
				refreshToken: testToken,
			},
			want: want{
				cookie:     true,
				body:       `{"Message":"Login successful"}`,
				statusCode: http.StatusOK,
			},
		},
		{
			name:     "Unauthorized",
			userJSON: `{"email":"pbsal@yaoo.com","password":"wrongPassword"}`,
			userRequest: usermodels.UserLoginRequest{
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: string(testPassHash),
			},
			req:        "/login",
			method:     http.MethodPost,
			mockFlagDB: true,
			err:        nil,
			want: want{
				cookie:     true,
				body:       `{"error":"the creds are invalid"}`,
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:     "Invalid JSON request body",
			userJSON: `{"email":123,"password":"wrongPassword"}`,
			userRequest: usermodels.UserLoginRequest{
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: string(testPassHash),
			},
			req:        "/login",
			method:     http.MethodPost,
			mockFlagDB: false,
			err:        nil,
			want: want{
				cookie:     true,
				body:       `{"error":"json: cannot unmarshal number into Go struct field UserLoginRequest.email of type string"}`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:     "Error from DB",
			userJSON: fmt.Sprintf(`{"email":"pbsal@yaoo.com","password":"%s"}`, testPass),
			userRequest: usermodels.UserLoginRequest{
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: string(testPassHash),
			},
			req:        "/login",
			method:     http.MethodPost,
			mockFlagDB: true,
			err:        context.DeadlineExceeded,
			want: want{
				cookie:     true,
				body:       `{"error":"context deadline exceeded"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name:     "Error from tokenSigner Access",
			userJSON: `{"email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userRequest: usermodels.UserLoginRequest{
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: string(testPassHash),
			},
			req:                       "/login",
			method:                    http.MethodPost,
			mockFlagDB:                true,
			err:                       nil,
			mockFlagTokenSignerAccess: true,
			TokenSignerResponse: TokenSignerResp{
				accessTokenError: errors.New("access token error"),
				refreshToken:     testToken,
			},
			want: want{
				cookie:     true,
				body:       `{"error":"access token error"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
		{
			name:     "Error from tokenSigner Refresh",
			userJSON: `{"email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userRequest: usermodels.UserLoginRequest{
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: string(testPassHash),
			},
			req:                        "/login",
			method:                     http.MethodPost,
			mockFlagDB:                 true,
			mockFlagTokenSignerAccess:  true,
			mockFlagTokenSignerRefresh: true,
			TokenSignerResponse: TokenSignerResp{
				accessToken:       testToken,
				refreshTokenError: errors.New("refresh token error"),
			},
			want: want{
				cookie:     true,
				body:       `{"error":"refresh token error"}`,
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewStorage(t)
			srv.db = repo
			jwtTokenSigner := mocks.NewTokenSigner(t)
			srv.tokenSigner = jwtTokenSigner
			if tc.mockFlagDB {
				repo.On("GetUserByEmail", tc.userRequest.Email).Return(tc.userFromDB, tc.err)
			}
			if tc.mockFlagTokenSignerAccess {
				jwtTokenSigner.On("NewAccessToken", mock.Anything).
					Return(tc.TokenSignerResponse.accessToken, tc.TokenSignerResponse.accessTokenError)
			}
			if tc.mockFlagTokenSignerRefresh {
				jwtTokenSigner.On("NewRefreshToken", mock.Anything).
					Return(tc.TokenSignerResponse.refreshToken, tc.TokenSignerResponse.refreshTokenError)
			}
			req := resty.New().R()
			req.URL = httpSrv.URL + tc.req
			req.Method = tc.method
			req.Body = tc.userJSON

			res, errSend := req.Send()

			require.NoError(t, errSend)
			assert.Equal(t, tc.want.statusCode, res.StatusCode())
			assert.Equal(t, tc.want.body, string(res.Body()))
			assert.NotNil(t, res.Cookies())
		})
	}
}

func TestUpdateUser(t *testing.T) {
	var srv ToDoListAPI
	gin.SetMode(gin.ReleaseMode)

	type want struct {
		body       string
		statusCode int
	}

	type test struct {
		name            string
		userJSON        string
		userFromDB      usermodels.User
		req             string
		method          string
		mockFlag        bool
		err             error
		errUpdate       error
		userIDFromParam string
		userIDFromCtx   string
		mockFlagCtx     bool
		want            want
	}

	tests := []test{
		{
			name:     "Update success",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:             "/users",
			method:          http.MethodPut,
			mockFlag:        true,
			err:             nil,
			userIDFromParam: "/testID",
			userIDFromCtx:   "testID",
			mockFlagCtx:     true,
			want: want{
				body:       `{"NewUserInfo":{"uuid":"2246b7cc-4afa-4e31-abc9-24f8c95692f1","name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}}`,
				statusCode: http.StatusOK,
			},
		},
		{
			name:     "User from context not exists",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:             "/users",
			method:          http.MethodPut,
			mockFlag:        false,
			err:             nil,
			userIDFromParam: "/testID",
			mockFlagCtx:     false,
			want: want{
				body:       `{"error":"unauthorized"}`,
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:     "User from context != from param",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:             "/users",
			method:          http.MethodPut,
			mockFlag:        false,
			err:             nil,
			userIDFromParam: "/fakeTestID",
			userIDFromCtx:   "testID",
			mockFlagCtx:     true,
			want: want{
				body:       `{"error":"unauthorized"}`,
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:     "Broken JSON",
			userJSON: `{"name":"pere","email": 1pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:             "/users",
			method:          http.MethodPut,
			mockFlag:        false,
			err:             nil,
			userIDFromParam: "/testID",
			userIDFromCtx:   "testID",
			mockFlagCtx:     true,
			want: want{
				body:       `"invalid character 'p' after object key:value pair"`,
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:     "Bad request from DB",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:             "/users",
			method:          http.MethodPut,
			mockFlag:        true,
			err:             nil,
			errUpdate:       errors.New("error from DB"),
			userIDFromParam: "/testID",
			userIDFromCtx:   "testID",
			mockFlagCtx:     true,
			want: want{
				body:       `"error from DB"`,
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			if tc.mockFlagCtx {
				r.Use(func(c *gin.Context) {
					c.Set("userID", tc.userIDFromCtx)
					c.Next()
				})
			}
			r.PUT("/users/:id", srv.updateUser)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			repo := mocks.NewStorage(t)
			srv.db = repo
			if tc.mockFlag {
				repo.On("GetUserByID", tc.userIDFromCtx).Return(tc.userFromDB, tc.err)
				repo.On("UpdateUser", mock.MatchedBy(func(user usermodels.User) bool {
					return user.Name == tc.userFromDB.Name && user.Email == tc.userFromDB.Email
				})).Return(tc.userFromDB, tc.errUpdate)
			}
			req := resty.New().R()
			req.URL = httpSrv.URL + tc.req + tc.userIDFromParam
			req.Method = tc.method
			req.Body = tc.userJSON

			res, err := req.Send()
			require.NoError(t, err)
			assert.Equal(t, tc.want.statusCode, res.StatusCode())
			if tc.err == nil {
				assert.Equal(t, tc.want.body, string(res.Body()))
			} else {
				assert.Contains(t, string(res.Body()), tc.want.body)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	var srv ToDoListAPI
	gin.SetMode(gin.ReleaseMode)

	type want struct {
		body       string
		statusCode int
	}

	type test struct {
		name            string
		userJSON        string
		userFromDB      usermodels.User
		req             string
		method          string
		mockFlag        bool
		err             error
		userIDFromParam string
		userIDFromCtx   string
		mockFlagCtx     bool
		want            want
	}

	tests := []test{
		{
			name:     "Delete success",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:             "/users",
			method:          http.MethodDelete,
			mockFlag:        true,
			err:             nil,
			userIDFromParam: "/testID",
			userIDFromCtx:   "testID",
			mockFlagCtx:     true,
			want: want{
				body:       `"User was deleted"`,
				statusCode: http.StatusOK,
			},
		},
		{
			name:     "User from context not exists",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:             "/users",
			method:          http.MethodDelete,
			mockFlag:        false,
			err:             nil,
			userIDFromParam: "/testID",
			mockFlagCtx:     false,
			want: want{
				body:       `{"error":"unauthorized"}`,
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:     "User from context != from param",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:             "/users",
			method:          http.MethodDelete,
			mockFlag:        false,
			err:             nil,
			userIDFromParam: "/fakeTestID",
			userIDFromCtx:   "testID",
			mockFlagCtx:     true,
			want: want{
				body:       `{"error":"unauthorized"}`,
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:     "Error from DB",
			userJSON: `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`,
			userFromDB: usermodels.User{
				UUID:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
				Name:     "pere",
				Email:    "pbsal@yaoo.com",
				Password: "unmarshall me",
			},
			req:             "/users",
			method:          http.MethodDelete,
			mockFlag:        true,
			err:             errors.New("error from DB"),
			userIDFromParam: "/testID",
			userIDFromCtx:   "testID",
			mockFlagCtx:     true,
			want: want{
				body:       `"error from DB"`,
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			if tc.mockFlagCtx {
				r.Use(func(c *gin.Context) {
					c.Set("userID", tc.userIDFromCtx)
					c.Next()
				})
			}
			r.DELETE("/users/:id", srv.deleteUser)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			repo := mocks.NewStorage(t)
			srv.db = repo
			if tc.mockFlag {
				repo.On("DeleteUser", tc.userIDFromCtx).Return(tc.err)
			}
			req := resty.New().R()
			req.URL = httpSrv.URL + tc.req + tc.userIDFromParam
			req.Method = tc.method
			req.Body = tc.userJSON

			res, err := req.Send()
			require.NoError(t, err)
			assert.Equal(t, tc.want.statusCode, res.StatusCode())
			if tc.err == nil {
				assert.Equal(t, tc.want.body, string(res.Body()))
			} else {
				assert.Contains(t, string(res.Body()), tc.want.body)
			}
		})
	}
}

func BenchmarkRegister(b *testing.B) {
	var srv ToDoListApi
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.POST("/register", srv.register)
	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	repo := mocks.NewStorage(b)

	repo.On("SaveUser", mock.Anything).Return(
		user_models.User{
			Uuid:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
			Name:     "pere",
			Email:    "pbsal@yaoo.com",
			Password: "unmarshall me",
		}, nil)

	srv.db = repo

	req := resty.New().R()
	req.URL = httpSrv.URL + "/register"
	req.Method = http.MethodPost
	req.Body = `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`

	for i := 0; i < b.N; i++ {
		_, _ = req.Send()
	}
}

func BenchmarkLogin(b *testing.B) {
	var srv ToDoListApi
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.POST("/login", srv.login)
	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	testPass := "unmarshall me"
	testPassHash, _ := bcrypt.GenerateFromPassword([]byte(testPass), bcrypt.DefaultCost)
	testToken := "any string"

	repo := mocks.NewStorage(b)
	repo.On("GetUserByEmail", mock.Anything).Return(
		user_models.User{
			Uuid:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
			Name:     "pere",
			Email:    "pbsal@yaoo.com",
			Password: string(testPassHash)},
		nil)
	srv.db = repo

	jwtTokenSigner := mocks.NewTokenSigner(b)

	jwtTokenSigner.On("NewAccessToken", mock.Anything).
		Return(testToken, nil)

	jwtTokenSigner.On("NewRefreshToken", mock.Anything).
		Return(testToken, nil)

	srv.tokenSigner = jwtTokenSigner

	req := resty.New().R()
	req.URL = httpSrv.URL + "/login"
	req.Method = http.MethodPost
	req.Body = fmt.Sprintf(`{"email":"pbsal@yaoo.com","password":"%s"}`, testPass)

	for i := 0; i < b.N; i++ {
		_, _ = req.Send()
	}
}

func BenchmarkUpdateUser(b *testing.B) {
	var srv ToDoListApi
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "testID")
		c.Next()
	},
	)

	r.PUT("/users/:id", srv.updateUser)

	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	repo := mocks.NewStorage(b)
	repo.On("GetUserByID", "testID").Return(
		user_models.User{
			Uuid:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
			Name:     "pere",
			Email:    "pbsal@yaoo.com",
			Password: "unmarshall me"},
		nil)
	repo.On("UpdateUser", mock.Anything).Return(
		user_models.User{
			Uuid:     "2246b7cc-4afa-4e31-abc9-24f8c95692f1",
			Name:     "pere",
			Email:    "pbsal@yaoo.com",
			Password: "unmarshall me"},
		nil)
	srv.db = repo

	req := resty.New().R()
	req.URL = httpSrv.URL + "/users" + "/testID"
	req.Method = http.MethodPut
	req.Body = `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`

	for i := 0; i < b.N; i++ {
		_, _ = req.Send()
	}
}

func BenchmarkDeleteUser(b *testing.B) {
	var srv ToDoListApi
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "testID")
		c.Next()
	})

	r.DELETE("/users/:id", srv.deleteUser)

	httpSrv := httptest.NewServer(r)
	defer httpSrv.Close()

	repo := mocks.NewStorage(b)

	repo.On("DeleteUser", "testID").Return(nil)

	srv.db = repo

	req := resty.New().R()
	req.URL = httpSrv.URL + "/users" + "/testID"
	req.Method = http.MethodDelete
	req.Body = `{"name":"pere","email":"pbsal@yaoo.com","password":"unmarshall me"}`

	for i := 0; i < b.N; i++ {
		_, _ = req.Send()
	}
}
