package userservice

import (
	"toDoList/internal/domain/user/usererrors"
	"toDoList/internal/domain/user/usermodels"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserStorage interface {
	GetAllUsers() ([]usermodels.User, error)
	SaveUser(user usermodels.User) (usermodels.User, error)
	GetUserByID(userID string) (usermodels.User, error)
	GetUserByEmail(email string) (usermodels.User, error)
	UpdateUser(user usermodels.User) (usermodels.User, error)
	DeleteUser(userID string) error
}

type UserService struct {
	db    UserStorage
	valid *validator.Validate
}

func NewUserService(db UserStorage) *UserService {
	return &UserService{db: db, valid: validator.New()}
}

func (us *UserService) GetAllUsers() ([]usermodels.User, error) {
	return us.db.GetAllUsers()
}

func (us *UserService) GetUserByID(userID string) (usermodels.User, error) {
	if userID == "" {
		return usermodels.User{}, usererrors.ErrUserEmptyInsert
	}
	user, err := us.db.GetUserByID(userID)
	if err != nil {
		return usermodels.User{}, err
	}
	return user, nil
}

func (us *UserService) SaveUser(newUser usermodels.UserRequest) (usermodels.User, error) {
	err := us.valid.Struct(newUser)
	if err != nil {
		return usermodels.User{}, err
	}

	var user usermodels.User

	uid := uuid.New().String()
	hash, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return usermodels.User{}, err
	}

	user.UUID = uid
	user.Name = newUser.Name
	user.Email = newUser.Email
	user.Password = string(hash)
	return us.db.SaveUser(user)
}

func (us *UserService) LoginUser(userReq usermodels.UserLoginRequest) (usermodels.User, error) {
	email := userReq.Email
	dbUser, err := us.db.GetUserByEmail(email)
	if err != nil {
		return usermodels.User{}, err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(userReq.Password)); err != nil {
		return usermodels.User{}, usererrors.ErrInvalidPassword
	}

	return dbUser, nil
}

func (us *UserService) UpdateUser(userID string, user usermodels.UserRequest) (usermodels.User, error) {
	err := us.valid.Struct(user)
	if err != nil {
		return usermodels.User{}, err
	}

	userInfo, err := us.db.GetUserByID(userID)
	if err != nil {
		return usermodels.User{}, err
	}

	userInfo.Name = user.Name
	userInfo.Email = user.Email
	userInfo.Password = user.Password

	newUserFullInfo, err := us.db.UpdateUser(userInfo)
	if err != nil {
		return usermodels.User{}, err
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
