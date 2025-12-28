package user_service

import (
	"toDoList/internal/domain/user/user_errors"
	"toDoList/internal/domain/user/user_models"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserStorage interface {
	GetAllUsers() ([]user_models.User, error)
	SaveUser(user user_models.User) (user_models.User, error)
	GetUserByID(userID string) (user_models.User, error)
	GetUserByEmail(email string) (user_models.User, error)
	UpdateUser(user user_models.User) (user_models.User, error)
	DeleteUser(userID string) error
}

type UserService struct {
	db    UserStorage
	valid *validator.Validate
}

func NewUserService(db UserStorage) *UserService {
	return &UserService{db: db, valid: validator.New()}
}

func (us *UserService) GetAllUsers() ([]user_models.User, error) {
	return us.db.GetAllUsers()
}

func (us *UserService) GetUserByID(userID string) (user_models.User, error) {
	if userID == "" {
		return user_models.User{}, user_errors.ErrorUserEmptyInsert
	}
	user, err := us.db.GetUserByID(userID)
	if err != nil {
		return user_models.User{}, err
	}
	return user, nil
}

func (us *UserService) SaveUser(newUser user_models.UserRequest) (user_models.User, error) {
	err := us.valid.Struct(newUser)
	if err != nil {
		return user_models.User{}, err
	}

	var user user_models.User

	uid := uuid.New().String()
	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return user_models.User{}, err
	}

	user.Uuid = uid
	user.Name = newUser.Name
	user.Email = newUser.Email
	user.Password = string(hash)
	return us.db.SaveUser(user)
}

func (us *UserService) LoginUser(userReq user_models.UserLoginRequest) (user_models.User, error) {
	email := userReq.Email
	dbUser, err := us.db.GetUserByEmail(email)
	if err != nil {
		return user_models.User{}, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(userReq.Password)); err != nil {
		return user_models.User{}, user_errors.ErrorInvalidPassword
	}

	return dbUser, nil
}

func (us *UserService) UpdateUser(userID string, user user_models.UserRequest) (user_models.User, error) {
	err := us.valid.Struct(user)
	if err != nil {
		return user_models.User{}, err
	}

	userInfo, err := us.db.GetUserByID(userID)
	if err != nil {
		return user_models.User{}, err
	}

	userInfo.Name = user.Name
	userInfo.Email = user.Email
	userInfo.Password = user.Password

	newUserFullInfo, err := us.db.UpdateUser(userInfo)
	if err != nil {
		return user_models.User{}, err
	}

	return newUserFullInfo, nil
}

func (us *UserService) DeleteUser(userID string) error {
	err := us.db.DeleteUser(userID)
	if err != nil {
		return err
	}
	return nil
}
